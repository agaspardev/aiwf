package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/agaspardev/aiwf/internal/workspace"
)

type scopeArgs struct {
	Subproject string
	Change     string
}

type resolvedScope struct {
	Subproject string
	Change     string
	Paths      workspace.Paths
}

func parseScopedArgs(args []string) (scopeArgs, []string, error) {
	var scope scopeArgs
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--subproject":
			if i+1 >= len(args) {
				return scopeArgs{}, nil, fmt.Errorf("--subproject requiere un valor")
			}
			scope.Subproject = args[i+1]
			i++
		case "--change":
			if i+1 >= len(args) {
				return scopeArgs{}, nil, fmt.Errorf("--change requiere un valor")
			}
			scope.Change = args[i+1]
			i++
		default:
			rest = append(rest, args[i])
		}
	}
	return scope, rest, nil
}

func resolveScope(root string, requested scopeArgs, requireChange bool) (resolvedScope, error) {
	subprojects, err := childDirectories(filepath.Join(root, ".ai-workflow", "projects"))
	if err != nil {
		return resolvedScope{}, err
	}
	subproject, err := workspace.ResolveIdentity(requested.Subproject, os.Getenv("AIWF_SUBPROJECT"), subprojects)
	if err != nil {
		return resolvedScope{}, fmt.Errorf("resolver subproject: %w", err)
	}

	change := ""
	if requireChange || requested.Change != "" || os.Getenv("AIWF_CHANGE") != "" {
		changes, listErr := childDirectories(filepath.Join(root, ".ai-workflow", "projects", subproject, "changes"))
		if listErr != nil {
			return resolvedScope{}, listErr
		}
		change, err = workspace.ResolveIdentity(requested.Change, os.Getenv("AIWF_CHANGE"), changes)
		if err != nil {
			return resolvedScope{}, fmt.Errorf("resolver change: %w", err)
		}
	}

	paths, err := workspace.NewPaths(root, subproject, change)
	if err != nil {
		return resolvedScope{}, err
	}
	return resolvedScope{Subproject: subproject, Change: change, Paths: paths}, nil
}

func childDirectories(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	children := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			children = append(children, entry.Name())
		}
	}
	sort.Strings(children)
	return children, nil
}
