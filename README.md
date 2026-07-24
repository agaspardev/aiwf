# aiwf — AI Workflow Framework

[![CI](https://github.com/agaspardev/aiwf/actions/workflows/ci.yml/badge.svg)](https://github.com/agaspardev/aiwf/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/agaspardev/aiwf)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**aiwf** is an AI coding agent installer and workflow orchestrator. Built on top of [Gentle-AI](https://github.com/Gentleman-Programming/gentle-ai), it provisions AI coding agents with Spec-Driven Development (SDD), persistent memory, curated skills, security pipelines, and SonarQube integration.

## Features

| Capability | Description |
|-----------|-------------|
| **Bootstrap** | Detects OS/arch, package managers, and runtime prereqs (Go, Node, Python, Docker, Git) — `aiwf doctor` |
| **Install & Update** | Installs gentle-ai + aiwf overlay; 2-week adoption window for new releases — `aiwf install` |
| **Overlay System** | Owned, MarkerBlock, and JSONMerge strategies for coexistence with gentle-ai — `aiwf install`, `aiwf reconcile`, `aiwf uninstall` |
| **SDD Orchestrator** | Full Spec-Driven Development workflow from proposal to archive — `aiwf sdd-*` |
| **Security Pipeline** | Gitleaks, Semgrep, OSV-Scanner, Trivy, Syft orchestration — `aiwf security` |
| **SonarQube** | Quality gate, issue tracking, code review integration — `aiwf sonar` |
| **Documentation** | Automatic project structure analysis and architecture docs — `aiwf document` |
| **Gatekeeper** | Quality gate enforcement (lint, typecheck, tests, 4R review) — `aiwf gate` |

## Quick start

```bash
# Install (requires gentle-ai)
aiwf install

# Check your setup
aiwf doctor

# Run a mode
aiwf prueba      # Development mode
aiwf test        # Test mode
aiwf production   # Production mode
```

## Requirements

- **Go 1.25+** for building from source
- **Gentle-AI** (detected automatically, installed via `aiwf install` if missing)
- Optional: SonarQube (for `aiwf sonar`), security scanners (for `aiwf security`)

## Installation

### From source

```bash
git clone https://github.com/agaspardev/aiwf.git
cd aiwf
go build -o aiwf ./cmd/aiwf
```

### Post-install

```bash
# Initialize a project workspace
aiwf init

# Verify everything works
aiwf doctor --verbose
```

## Commands

```
aiwf <mode>          Run a development mode (prueba, test, production)
aiwf doctor          System diagnostics and prereq checks
aiwf install         Install gentle-ai + aiwf overlay
aiwf reconcile       Re-apply overlay without reinstalling
aiwf uninstall       Remove aiwf overlay (gentle-ai untouched)
aiwf init            Initialize project workspace structure
aiwf gate            Run quality gate checks
aiwf skills          List installed skills
aiwf security        Security scanning pipeline
aiwf sonar           SonarQube integration
aiwf document        Project documentation generation
aiwf gemini          Consult Gemini for second opinions
aiwf estado          Show current state
aiwf diagnostico     Deep diagnostics
```

## Architecture

```
cmd/aiwf/            CLI entry point (main, satellites, run, ops)
internal/
├── assets/          Embedded files (skills, rules, templates, agents, security)
├── bootstrap/       OS detection, prereqs, package managers
├── config/          Runtime configuration, paths
├── diag/            Diagnostics
├── docgen/          Automatic project documentation
├── gatekeeper/      Quality gate enforcement
├── harness/         Mode resolution, contract generation, Claude Code args
├── initproj/        Project initialization
├── omniroute/       OmniRoute API integration (Gemini consultation)
├── overlay/         File management: Owned, MarkerBlock, JSONMerge
├── security/        Gitleaks, Semgrep, OSV, Trivy, Syft orchestration
├── skillreg/        Skill registry management
├── sonar/           SonarQube API integration
├── state/           State management
└── upstream/        Gentle-AI detection, releases, installation, advisories
```

## Dependencies

aiwf minimizes external dependencies. The only third-party dependency is:

- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) (MIT + Apache 2.0)

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for full license details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE) © 2026 Antonio Gaspar
