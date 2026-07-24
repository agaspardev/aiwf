# Skill: /aiwf-osint [dep|cve|vendor|license]

OSINT defensivo documentado. Solo opera sobre targets autorizados explícitamente.

## Subcomandos

```
/aiwf-osint dep      # analiza árbol de dependencias del proyecto
/aiwf-osint cve      # busca CVEs en dependencias (OSV-Scanner)
/aiwf-osint vendor   # reputación y actividad de vendors de dependencias clave
/aiwf-osint license  # análisis de licencias (compatibilidad, copyleft)
```

## Restricciones de seguridad (SIEMPRE)

1. **Verificar `.security/policies/osint-allowlist.yaml`** antes de cualquier reconocimiento externo.
2. Si el subcomando requiere acceso a sistemas externos (CVE databases, package registries): solo queries de LECTURA, nunca write.
3. Para reconocimiento activo de dominios (`staging`, `nightly`): RECHAZAR si no hay `authorized_domains` en `osint-allowlist.yaml`.
4. Nunca consultar dominios de terceros sin autorización escrita del propietario.
5. Reportar qué herramientas se usaron y contra qué targets.

## Comportamiento por subcomando

### dep
Analizar `package.json`, `go.mod`, `requirements.txt`, `Cargo.toml` y producir árbol de dependencias con:
- Total de dependencias directas e indirectas
- Dependencias con acceso a filesystem, network, o exec
- Dependencias no mantenidas (último release > 2 años)

### cve
Ejecutar `ai security sca` (OSV-Scanner + Trivy) y reportar hallazgos críticos.

### vendor
Para las 5 dependencias más críticas: verificar en npm/PyPI/pkg.go.dev:
- Mantenimiento activo (releases en últimos 6 meses)
- Número de mantenedores (riesgo de bus factor)
- Descarga semanal (popularidad = mayor escrutinio comunitario)

### license
Analizar licencias del árbol de dependencias y reportar:
- Licencias copyleft (GPL, AGPL) que pueden afectar distribución
- Incompatibilidades conocidas
- Dependencias sin licencia declarada
