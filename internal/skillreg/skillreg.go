// Package skillreg genera el índice de skills (registry) y detecta drift, portado de
// skill-registry.ps1. Data de ruteo determinista (cero tokens): el orquestador lee el
// registry y rutea por trigger.
package skillreg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry es una skill en el registry.
type Entry struct {
	Name    string
	Trigger string
	Path    string // relativo a skillsDir, con separador '/'
}

// Scan recolecta skills desde skillsDir: todos los SKILL.md, más aiwf/*.md y rules/*.md.
func Scan(skillsDir string) ([]Entry, error) {
	var paths []string

	// SKILL.md recursivo.
	_ = filepath.WalkDir(skillsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == "SKILL.md" {
			paths = append(paths, p)
		}
		return nil
	})
	// rules/*.md (legacy flat format, kept for rule modules).
	matches, _ := filepath.Glob(filepath.Join(skillsDir, "rules", "*.md"))
	paths = append(paths, matches...)

	var entries []Entry
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		e := entryFromFile(skillsDir, p, string(data))
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

func entryFromFile(skillsDir, path, content string) Entry {
	fm := parseFrontmatter(content)

	name := fm["name"]
	if name == "" {
		if filepath.Base(path) == "SKILL.md" {
			name = filepath.Base(filepath.Dir(path))
		} else {
			name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
	}

	trigger := firstNonEmpty(fm["trigger"], fm["description"], firstH1(content))
	if len(trigger) > 160 {
		trigger = trigger[:157] + "..."
	}

	rel, err := filepath.Rel(skillsDir, path)
	if err != nil {
		rel = path
	}
	return Entry{Name: name, Trigger: trigger, Path: filepath.ToSlash(rel)}
}

// parseFrontmatter lee el bloque YAML `--- ... ---` inicial como pares key: value.
func parseFrontmatter(content string) map[string]string {
	fm := map[string]string{}
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			key := strings.TrimSpace(k)
			if key == "" || strings.HasPrefix(key, " ") {
				continue
			}
			fm[key] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return fm
}

func firstH1(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// maxDescriptionLines is the hard budget for skill descriptions.
// Only eagerly loaded content; keep it short.
const maxDescriptionLines = 2

// PackagingError describes a single validation failure in a skill package.
type PackagingError struct {
	Path    string // relative to skills dir
	Problem string
}

func (e PackagingError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Problem)
}

// ValidatePackaging checks that every SKILL.md under skillsDir follows the
// packaging contract:
//   - lives inside its own directory (aiwf-<name>/SKILL.md)
//   - has frontmatter with "name" matching the directory name
//   - has a "description" that is non-empty and at most maxDescriptionLines
//
// Legacy flat files (aiwf/*.md, rules/*.md) are skipped — they are not part
// of the new packaging contract and will be removed by the migration task.
func ValidatePackaging(skillsDir string) []PackagingError {
	var errs []PackagingError

	_ = filepath.WalkDir(skillsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			errs = append(errs, PackagingError{Path: p, Problem: "cannot read file"})
			return nil
		}

		rel, _ := filepath.Rel(skillsDir, p)
		relSlash := filepath.ToSlash(rel)
		dirName := filepath.Base(filepath.Dir(p))
		fm := parseFrontmatter(string(data))

		// name must exist
		name := fm["name"]
		if name == "" {
			errs = append(errs, PackagingError{Path: relSlash, Problem: "missing frontmatter 'name'"})
		} else if name != dirName {
			errs = append(errs, PackagingError{
				Path:    relSlash,
				Problem: fmt.Sprintf("name %q does not match directory %q", name, dirName),
			})
		}

		// description must exist and respect budget
		desc := fm["description"]
		if desc == "" {
			errs = append(errs, PackagingError{Path: relSlash, Problem: "missing frontmatter 'description'"})
		} else {
			lines := strings.Split(strings.TrimSpace(desc), "\n")
			if len(lines) > maxDescriptionLines {
				errs = append(errs, PackagingError{
					Path: relSlash,
					Problem: fmt.Sprintf("description has %d lines, max %d",
						len(lines), maxDescriptionLines),
				})
			}
		}

		return nil
	})

	return errs
}

// Lint detecta drift: skills sin trigger y nombres duplicados.
func Lint(entries []Entry) []string {
	var problems []string
	counts := map[string]int{}
	for _, e := range entries {
		if strings.TrimSpace(e.Trigger) == "" {
			problems = append(problems, "sin trigger/descripcion: "+e.Name)
		}
		counts[e.Name]++
	}
	for name, n := range counts {
		if n > 1 {
			problems = append(problems, fmt.Sprintf("nombre duplicado: %s (%d veces)", name, n))
		}
	}
	sort.Strings(problems)
	return problems
}

// Generate produce el registry en markdown.
func Generate(entries []Entry) string {
	var b strings.Builder
	b.WriteString("# Skill Registry (GENERADO por aiwf skills — no editar a mano)\n\n")
	fmt.Fprintf(&b, "Skills: **%d**. El orquestador lee este indice y rutea por trigger antes de ejecutar.\n\n", len(entries))
	b.WriteString("| Skill | Trigger / cuando usar | Path |\n|---|---|---|\n")
	for _, e := range entries {
		t := strings.ReplaceAll(e.Trigger, "|", `\|`)
		fmt.Fprintf(&b, "| %s | %s | `%s` |\n", e.Name, t, e.Path)
	}
	return b.String()
}
