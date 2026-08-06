// Package docgen documenta un proyecto de forma DETERMINISTA (cero tokens): detecta
// stack, estructura, dependencias, entrypoints, conexiones externas y metadatos git, y
// produce context-pack.md + ARCHITECTURE.md + un reporte crudo JSON. Portado de
// document-project.ps1. La síntesis con LLM es opt-in y SIEMPRE vía OmniRoute.
package docgen

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/agaspardev/aiwf/internal/omniroute"
)

// ExtCount / DirCount son conteos ordenables.
type ExtCount struct {
	Ext   string `json:"ext"`
	Count int    `json:"count"`
}
type DirCount struct {
	Dir   string `json:"dir"`
	Count int    `json:"count"`
}

// Report es el reporte crudo del proyecto.
type Report struct {
	GeneratedAt  string            `json:"generated_at"`
	Project      string            `json:"project"`
	Stack        []string          `json:"stack"`
	FilesTotal   int               `json:"files_total"`
	TopDirs      []DirCount        `json:"top_dirs"`
	ByExtension  []ExtCount        `json:"by_extension"`
	Dependencies []string          `json:"dependencies"`
	DepsTotal    int               `json:"deps_total"`
	Entrypoints  []string          `json:"entrypoints"`
	External     map[string]int    `json:"external"`
	CodeGraph    string            `json:"codegraph"`
	Git          map[string]string `json:"git"`
}

// Result reporta los archivos generados.
type Result struct {
	ReportPath       string
	ContextPackPath  string
	ArchitecturePath string
	ArchivedPrev     string // versión previa archivada (update mode), si hubo cambio
}

// Generate ejecuta el análisis y escribe conocimiento durable y evidencia en
// destinos explícitos. El caller resuelve ownership; docgen no infiere contexto.
func Generate(root, knowledge, evidence, mode string, synthesize bool) (*Report, *Result, error) {
	if knowledge == "" || evidence == "" {
		return nil, nil, fmt.Errorf("knowledge y evidence son obligatorios")
	}
	history := filepath.Join(knowledge, "history")
	documentEvidence := filepath.Join(evidence, "document")
	for _, d := range []string{knowledge, history, documentEvidence} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, nil, err
		}
	}
	ts := time.Now().Format("20060102-150405")
	today := time.Now().Format("2006-01-02")

	files, err := collectFiles(root)
	if err != nil {
		return nil, nil, fmt.Errorf("collecting files: %w", err)
	}
	rep := &Report{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Project:      filepath.Base(root),
		Stack:        detectStack(root),
		FilesTotal:   len(files),
		TopDirs:      topDirs(files, 20),
		ByExtension:  byExtension(files, 15),
		Dependencies: parseDeps(root),
		Entrypoints:  entrypoints(root),
		External:     scanConnections(root, files),
		CodeGraph:    codegraphNote(root),
		Git:          gitInfo(root),
	}
	rep.DepsTotal = len(rep.Dependencies)
	if len(rep.Dependencies) > 200 {
		rep.Dependencies = rep.Dependencies[:200]
	}

	res := &Result{}

	// Reporte crudo.
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return rep, nil, err
	}
	res.ReportPath = filepath.Join(documentEvidence, "report-"+ts+".json")
	if err := os.WriteFile(res.ReportPath, data, 0o644); err != nil {
		return rep, nil, err
	}

	// context-pack.md
	res.ContextPackPath = filepath.Join(knowledge, "context-pack.md")
	if err := os.WriteFile(res.ContextPackPath, []byte(contextPack(rep, today)), 0o644); err != nil {
		return rep, res, err
	}

	// ARCHITECTURE.md con control de evolución.
	res.ArchitecturePath = filepath.Join(knowledge, "ARCHITECTURE.md")
	newArch := architecture(rep, today, ts)
	if mode == "update" {
		if prev, err := os.ReadFile(res.ArchitecturePath); err == nil {
			if stripFrontmatter(string(prev)) != stripFrontmatter(newArch) {
				archived := filepath.Join(history, "ARCHITECTURE-"+ts+".md")
				if err := os.WriteFile(archived, prev, 0o644); err != nil {
					return rep, res, fmt.Errorf("archiving previous ARCHITECTURE.md: %w", err)
				}
				newArch = strings.Replace(newArch, "supersedes: []",
					"supersedes: [history/ARCHITECTURE-"+ts+".md]", 1)
				res.ArchivedPrev = archived
			}
		}
	}
	if err := os.WriteFile(res.ArchitecturePath, []byte(newArch), 0o644); err != nil {
		return rep, res, err
	}

	// Síntesis opt-in (vía OmniRoute, nunca IA local).
	if synthesize {
		if prose, err := synthesizeProse(rep); err == nil && prose != "" {
			f, err := os.OpenFile(res.ArchitecturePath, os.O_WRONLY|os.O_APPEND, 0o644)
			if err == nil {
				fmt.Fprintf(f, "\n## Síntesis (IA via OmniRoute — no vinculante, verificar)\n\n%s\n", prose)
				f.Close()
			}
		}
	}
	return rep, res, nil
}

// ── analizadores (puros/testeables) ──────────────────────────────────────────

var stackManifests = []struct {
	name  string
	globs []string
}{
	{"Node/JS-TS", []string{"package.json"}},
	{"Go", []string{"go.mod"}},
	{"Python", []string{"requirements.txt", "pyproject.toml"}},
	{"Rust", []string{"Cargo.toml"}},
	{".NET", []string{"*.csproj", "*.sln"}},
	{"Java/Maven", []string{"pom.xml"}},
	{"Java/Gradle", []string{"build.gradle"}},
}

func detectStack(root string) []string {
	var detected []string
	for _, m := range stackManifests {
		for _, g := range m.globs {
			matches, _ := filepath.Glob(filepath.Join(root, g))
			if len(matches) > 0 {
				detected = append(detected, m.name)
				break
			}
		}
	}
	if len(detected) == 0 {
		return []string{"desconocido"}
	}
	return detected
}

var excludedDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true,
	".ai-workflow": true, "coverage": true, "target": true, "bin": true, "obj": true,
}

func collectFiles(root string) ([]string, error) {
	if isGitRepo(root) {
		if out := runGit(root, "ls-files"); out != "" {
			return strings.Split(out, "\n"), nil
		}
	}
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != root && excludedDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func byExtension(files []string, top int) []ExtCount {
	counts := map[string]int{}
	for _, f := range files {
		if ext := filepath.Ext(f); ext != "" {
			counts[ext]++
		}
	}
	return topCounts(counts, top, func(k string, v int) ExtCount { return ExtCount{Ext: k, Count: v} })
}

func topDirs(files []string, top int) []DirCount {
	counts := map[string]int{}
	for _, f := range files {
		if i := strings.IndexAny(f, "/\\"); i > 0 {
			counts[f[:i]]++
		}
	}
	return topCounts(counts, top, func(k string, v int) DirCount { return DirCount{Dir: k, Count: v} })
}

// topCounts ordena un mapa por conteo desc (desempate por clave) y toma los primeros n.
func topCounts[T any](counts map[string]int, n int, mk func(string, int) T) []T {
	type kv struct {
		k string
		v int
	}
	pairs := make([]kv, 0, len(counts))
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v != pairs[j].v {
			return pairs[i].v > pairs[j].v
		}
		return pairs[i].k < pairs[j].k
	})
	if len(pairs) > n {
		pairs = pairs[:n]
	}
	out := make([]T, len(pairs))
	for i, p := range pairs {
		out[i] = mk(p.k, p.v)
	}
	return out
}

type pkgJSON struct {
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPkg(root string) (pkgJSON, bool) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return pkgJSON{}, false
	}
	var p pkgJSON
	if json.Unmarshal(data, &p) != nil {
		return pkgJSON{}, false
	}
	return p, true
}

func parseDeps(root string) []string {
	var deps []string
	if p, ok := readPkg(root); ok {
		for _, sec := range []map[string]string{p.Dependencies, p.DevDependencies} {
			for _, name := range sortedKeys(sec) {
				deps = append(deps, name+"@"+sec[name])
			}
		}
	}
	if lines, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		// Cubre tanto el bloque `require ( ... )` (líneas indentadas) como la forma de
		// una sola línea `require mod vX.Y.Z`.
		re := regexp.MustCompile(`^[\w./-]+\s+v\d`)
		for _, l := range strings.Split(string(lines), "\n") {
			t := strings.TrimPrefix(strings.TrimSpace(l), "require ")
			if re.MatchString(t) {
				deps = append(deps, t)
			}
		}
	}
	if lines, err := os.ReadFile(filepath.Join(root, "requirements.txt")); err == nil {
		for _, l := range strings.Split(string(lines), "\n") {
			t := strings.TrimSpace(l)
			if t != "" && !strings.HasPrefix(t, "#") {
				deps = append(deps, t)
			}
		}
	}
	return deps
}

func entrypoints(root string) []string {
	var eps []string
	if p, ok := readPkg(root); ok {
		if p.Main != "" {
			eps = append(eps, "main: "+p.Main)
		}
		for _, name := range sortedKeys(p.Scripts) {
			eps = append(eps, "script:"+name)
		}
	}
	for _, cand := range []string{"main.go", "src/main.ts", "src/index.ts", "src/main.py", "app.py", "main.py", "Program.cs"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(cand))); err == nil {
			eps = append(eps, cand)
		}
	}
	return eps
}

var connPatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"HTTP endpoints", regexp.MustCompile(`https?://[a-zA-Z0-9.-]+`)},
	{"Postgres", regexp.MustCompile(`postgres(ql)?://`)},
	{"MySQL", regexp.MustCompile(`mysql://`)},
	{"MongoDB", regexp.MustCompile(`mongodb(\+srv)?://`)},
	{"Redis", regexp.MustCompile(`redis://`)},
	{"AWS", regexp.MustCompile(`\b(s3|dynamodb|sqs|lambda)\b`)},
	{"Env vars", regexp.MustCompile(`process\.env\.\w+|os\.environ|Environment\.GetEnvironmentVariable`)},
}

var scanExtRe = regexp.MustCompile(`\.(ts|js|go|py|cs|java|rb|php|ya?ml|json|toml)$`)

func scanConnections(root string, files []string) map[string]int {
	hits := map[string]int{}
	scanned := 0
	for _, rel := range files {
		if scanned >= 400 {
			break
		}
		if !scanExtRe.MatchString(rel) {
			continue
		}
		p := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Stat(p)
		if err != nil || info.Size() > 500*1024 {
			continue
		}
		scanned++
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, pat := range connPatterns {
			if n := len(pat.re.FindAll(data, -1)); n > 0 {
				hits[pat.name] += n
			}
		}
	}
	return hits
}

func codegraphNote(root string) string {
	if _, err := os.Stat(filepath.Join(root, ".codegraph")); err == nil {
		return "indexado (.codegraph/) — usar codegraph_explore para navegación de símbolos"
	}
	return "no indexado"
}

func gitInfo(root string) map[string]string {
	if !isGitRepo(root) {
		return map[string]string{}
	}
	info := map[string]string{
		"branch":     runGit(root, "rev-parse", "--abbrev-ref", "HEAD"),
		"commits":    runGit(root, "rev-list", "--count", "HEAD"),
		"lastCommit": runGit(root, "log", "-1", "--format=%ci"),
	}
	if sl := runGit(root, "shortlog", "-sn", "HEAD"); sl != "" {
		info["contributors"] = fmt.Sprintf("%d", len(strings.Split(sl, "\n")))
	}
	return info
}

// ── síntesis opt-in ──────────────────────────────────────────────────────────

func synthesizeProse(rep *Report) (string, error) {
	data, _ := json.MarshalIndent(rep, "", "  ")
	prompt := "Sos un arquitecto. A partir de este reporte determinista JSON de un proyecto, " +
		"escribi 2-3 parrafos claros sobre que hace el sistema, como se conecta y su arquitectura probable. " +
		"No inventes; si algo no esta en los datos, decilo. Datos:\n" + string(data)
	return omniroute.Consult(context.Background(), prompt, "agent-daily", 2048)
}

// ── utilidades ───────────────────────────────────────────────────────────────

func isGitRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func runGit(root string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var frontmatterRe = regexp.MustCompile(`(?s)^---.*?---`)

func stripFrontmatter(s string) string {
	return strings.TrimSpace(frontmatterRe.ReplaceAllString(s, ""))
}
