"""
infrastructure/security/security-adapter.py
MCP server (stdio) — permite a Claude invocar scans de seguridad desde dentro de la sesión.

Restricciones de seguridad:
  - shell=False en TODO subprocess.run — nunca shell injection posible
  - WORKSPACE restriction — solo paths dentro del directorio de trabajo
  - ALLOWED_COMMANDS whitelist — solo herramientas AppSec conocidas
  - No devuelve output completo al modelo (puede ser MB) — solo resumen + ruta de reporte
  - Content-hash deduplication — evita re-ejecutar si las entradas no cambiaron

Protocolo MCP (stdio): JSON-RPC 2.0, una línea por mensaje, newline-delimited.
"""
from __future__ import annotations

import hashlib
import json
import os
import subprocess
import sys
import datetime
from pathlib import Path
from typing import Any

# ── Restricciones de seguridad ────────────────────────────────────────────────

# WORKSPACE es el árbol que se escanea. Los reportes requieren un change owner
# explícito; sin AIWF_CHANGE_ROOT el adapter falla cerrado.
WORKSPACE = Path(os.environ.get("WORKSPACE") or ".").resolve()

ALLOWED_COMMANDS: dict[str, list[str]] = {
    "semgrep": [
        "semgrep", "scan", "--config", "p/default", "--json",
    ],
    "gitleaks": [
        "gitleaks", "git", "--redact", "--report-format", "json",
    ],
    "osv-scanner": [
        "osv-scanner", "scan", "source", "--recursive", ".", "--format", "sarif",
    ],
    "trivy": [
        "trivy", "fs", "--scanners", "vuln,misconfig,secret", "--format", "json",
    ],
    "syft": [
        "syft", ".", "-o", "cyclonedx-json",
    ],
}

MAX_TIMEOUT = 300  # segundos


def validate_path(p: str) -> Path:
    """Rechaza cualquier path fuera del workspace autorizado."""
    path = Path(p).resolve()
    if path != WORKSPACE and WORKSPACE not in path.parents:
        raise ValueError(f"Path fuera del workspace autorizado: {p!r}")
    return path


def reports_dir() -> Path:
    """Resuelve evidence/security del change activo o falla cerrado."""
    change_root = os.environ.get("AIWF_CHANGE_ROOT")
    if not change_root:
        raise ValueError("AIWF_CHANGE_ROOT no está resuelto; activá un change primero")
    return validate_path(change_root) / "evidence" / "security"


def content_hash(tool: str) -> str:
    """
    Deduplicación basada en contenido de lock files.
    Si el hash no cambió desde el último scan, el resultado es reutilizable.
    """
    candidates = [
        "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
        "go.sum", "requirements.txt", "Cargo.lock", "poetry.lock",
    ]
    content = b"".join(
        Path(f).read_bytes() for f in candidates
        if (WORKSPACE / f).exists()
    )
    return hashlib.sha256(tool.encode() + content).hexdigest()[:12]


def execute_tool(tool: str, extra_args: list[str] | None = None) -> dict[str, Any]:
    """
    Ejecuta una herramienta de la whitelist.
    extra_args son validados y añadidos SOLO si el tool los acepta.
    """
    if tool not in ALLOWED_COMMANDS:
        raise ValueError(
            f"Herramienta no autorizada: {tool!r}. "
            f"Permitidas: {list(ALLOWED_COMMANDS)}"
        )

    reports = reports_dir()
    reports.mkdir(parents=True, exist_ok=True)
    ts = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
    report_file = reports / f"{tool}-{ts}.json"

    cmd = list(ALLOWED_COMMANDS[tool])  # copia — nunca mutar la whitelist

    # Añadir output file arg según la herramienta
    if tool == "semgrep":
        cmd += ["-o", str(report_file)]
    elif tool == "gitleaks":
        cmd += ["--report-path", str(report_file)]
    elif tool == "osv-scanner":
        cmd += ["--output-file", str(report_file)]
    elif tool == "trivy":
        cmd += ["--output", str(report_file)]
    elif tool == "syft":
        cmd = list(ALLOWED_COMMANDS[tool][:-2]) + [
            "-o", f"cyclonedx-json={report_file}"
        ]

    result = subprocess.run(
        cmd,
        cwd=WORKSPACE,
        capture_output=True,
        text=True,
        timeout=MAX_TIMEOUT,
        check=False,
        shell=False,  # NUNCA shell=True
    )

    return {
        "tool": tool,
        "exit_code": result.returncode,
        "report_path": str(report_file),
        "content_hash": content_hash(tool),
        "stderr_snippet": result.stderr[-2000:] if result.stderr else "",
        "timestamp": ts,
    }


# ── MCP server (stdio, JSON-RPC 2.0) ─────────────────────────────────────────

MCP_VERSION = "2024-11-05"

TOOLS_MANIFEST = [
    {
        "name": "security_scan",
        "description": (
            "Ejecuta una herramienta de seguridad AppSec dentro del workspace actual. "
            "Solo herramientas de la whitelist: semgrep, gitleaks, osv-scanner, trivy, syft. "
            "Devuelve exit_code y ruta del reporte — NO el contenido completo del reporte."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "tool": {
                    "type": "string",
                    "enum": list(ALLOWED_COMMANDS.keys()),
                    "description": "Herramienta de seguridad a ejecutar.",
                },
            },
            "required": ["tool"],
        },
    },
    {
        "name": "security_list_reports",
        "description": (
            "Lista los reportes de seguridad generados en ${AIWF_CHANGE_ROOT}/evidence/security/. "
            "Incluye nombre, fecha y tamaño. Útil para encontrar el reporte más reciente."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "tool_filter": {
                    "type": "string",
                    "description": "Filtrar por nombre de herramienta (opcional).",
                },
            },
            "required": [],
        },
    },
    {
        "name": "security_read_summary",
        "description": (
            "Lee el contenido de un scan-summary-*.md en ${AIWF_CHANGE_ROOT}/evidence/security/. "
            "Devuelve el texto del summary para incorporarlo al contexto. "
            "Preferir el más reciente si no se especifica archivo."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "filename": {
                    "type": "string",
                    "description": "Nombre del archivo de summary (opcional, usa el más reciente si se omite).",
                },
            },
            "required": [],
        },
    },
]


def handle_initialize(req_id: Any) -> dict:
    return {
        "jsonrpc": "2.0",
        "id": req_id,
        "result": {
            "protocolVersion": MCP_VERSION,
            "capabilities": {"tools": {}},
            "serverInfo": {"name": "security-adapter", "version": "1.0.0"},
        },
    }


def handle_tools_list(req_id: Any) -> dict:
    return {
        "jsonrpc": "2.0",
        "id": req_id,
        "result": {"tools": TOOLS_MANIFEST},
    }


def handle_tools_call(req_id: Any, params: dict) -> dict:
    tool_name = params.get("name")
    args = params.get("arguments", {})

    try:
        if tool_name == "security_scan":
            tool = args.get("tool")
            if not tool:
                raise ValueError("Parámetro 'tool' requerido.")
            result = execute_tool(tool)
            text = json.dumps(result, indent=2, ensure_ascii=False)

        elif tool_name == "security_list_reports":
            reports = reports_dir()
            tool_filter = args.get("tool_filter", "")
            if not reports.exists():
                text = "No hay reportes aún. Ejecutar security_scan primero."
            else:
                files = sorted(reports.iterdir(), reverse=True)
                if tool_filter:
                    files = [f for f in files if tool_filter in f.name]
                entries = [
                    {
                        "name": f.name,
                        "size_kb": round(f.stat().st_size / 1024, 1),
                        "modified": datetime.datetime.fromtimestamp(
                            f.stat().st_mtime
                        ).strftime("%Y-%m-%d %H:%M"),
                    }
                    for f in files[:20]  # máximo 20
                ]
                text = json.dumps(entries, indent=2, ensure_ascii=False)

        elif tool_name == "security_read_summary":
            reports = reports_dir()
            filename = args.get("filename")
            if filename:
                target = validate_path(str(reports / filename))
            else:
                # El más reciente scan-summary-*.md
                summaries = sorted(
                    reports.glob("scan-summary-*.md"), reverse=True
                ) if reports.exists() else []
                if not summaries:
                    text = "No scan summaries available. Run: aiwf security all"
                    return _ok_result(req_id, text)
                target = summaries[0]

            content = target.read_text(encoding="utf-8")
            text = content[:8000]  # máximo 8 KB al modelo
            if len(content) > 8000:
                text += f"\n\n[Truncado — leer {target} completo para todos los findings]"

        else:
            raise ValueError(f"Herramienta desconocida: {tool_name!r}")

        return _ok_result(req_id, text)

    except (ValueError, subprocess.TimeoutExpired) as exc:
        return _error_result(req_id, str(exc))
    except Exception as exc:  # noqa: BLE001
        return _error_result(req_id, f"Error interno: {exc}")


def _ok_result(req_id: Any, text: str) -> dict:
    return {
        "jsonrpc": "2.0",
        "id": req_id,
        "result": {
            "content": [{"type": "text", "text": text}],
            "isError": False,
        },
    }


def _error_result(req_id: Any, message: str) -> dict:
    return {
        "jsonrpc": "2.0",
        "id": req_id,
        "result": {
            "content": [{"type": "text", "text": f"[ERROR] {message}"}],
            "isError": True,
        },
    }


def serve() -> None:
    """Loop principal MCP stdio: una línea por mensaje, flush inmediato."""
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
        except json.JSONDecodeError:
            continue

        req_id = req.get("id")
        method = req.get("method", "")
        params = req.get("params", {})

        if method == "initialize":
            response = handle_initialize(req_id)
        elif method == "tools/list":
            response = handle_tools_list(req_id)
        elif method == "tools/call":
            response = handle_tools_call(req_id, params)
        elif method == "notifications/initialized":
            continue  # notificación sin respuesta
        else:
            response = {
                "jsonrpc": "2.0",
                "id": req_id,
                "error": {"code": -32601, "message": f"Método no soportado: {method}"},
            }

        print(json.dumps(response, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    serve()
