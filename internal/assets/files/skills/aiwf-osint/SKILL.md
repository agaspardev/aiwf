---
name: aiwf-osint
description: Defensive OSINT for software supply chain — CVEs, deps, vendors, licenses.
---
# Skill: /aiwf-osint [dep|cve|vendor|license]

Documented defensive OSINT. Only operates on explicitly authorized targets.

## Subcommands

```
/aiwf-osint dep      # analyze project dependency tree
/aiwf-osint cve      # search CVEs in dependencies (OSV-Scanner)
/aiwf-osint vendor   # reputation and activity of key dependency vendors
/aiwf-osint license  # license analysis (compatibility, copyleft)
```

## Security constraints (ALWAYS)

1. **Check `.security/policies/osint-allowlist.yaml`** before any external reconnaissance.
2. If the subcommand requires access to external systems (CVE databases, package registries): read-only queries only, never write.
3. For active domain reconnaissance (`staging`, `nightly`): REJECT if no `authorized_domains` in `osint-allowlist.yaml`.
4. Never query third-party domains without written authorization from the owner.
5. Report which tools were used and against which targets.

## Behavior by subcommand

### dep
Analyze `package.json`, `go.mod`, `requirements.txt`, `Cargo.toml` and produce a dependency tree with:
- Total direct and indirect dependencies
- Dependencies with filesystem, network, or exec access
- Unmaintained dependencies (last release > 2 years)

### cve
Run `aiwf security sca` (OSV-Scanner + Trivy) and report critical findings.

### vendor
For the 5 most critical dependencies: check on npm/PyPI/pkg.go.dev:
- Active maintenance (releases in the last 6 months)
- Number of maintainers (bus factor risk)
- Weekly downloads (popularity = greater community scrutiny)

### license
Analyze the dependency tree licenses and report:
- Copyleft licenses (GPL, AGPL) that may affect distribution
- Known incompatibilities
- Dependencies without a declared license
