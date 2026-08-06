// Package workspace defines the canonical identities and paths of an aiwf workspace.
package workspace

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrInvalidIdentity   = errors.New("invalid workspace identity")
	ErrIdentityMissing   = errors.New("workspace identity is missing")
	ErrIdentityAmbiguous = errors.New("workspace identity is ambiguous")
)

// ValidateID accepts lowercase kebab-case identifiers only.
func ValidateID(value string) error {
	if value == "" {
		return fmt.Errorf("%w: value is empty", ErrInvalidIdentity)
	}
	if value[0] == '-' || value[len(value)-1] == '-' || strings.Contains(value, "--") {
		return fmt.Errorf("%w: %q is not kebab-case", ErrInvalidIdentity, value)
	}
	for _, r := range value {
		if r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return fmt.Errorf("%w: %q contains %q", ErrInvalidIdentity, value, r)
	}
	return nil
}

// ResolveIdentity resolves an identity without consulting global process state.
// Priority is explicit value, session value, then a unique candidate.
func ResolveIdentity(explicit, session string, candidates []string) (string, error) {
	for _, candidate := range candidates {
		if err := ValidateID(candidate); err != nil {
			return "", fmt.Errorf("candidate %q: %w", candidate, err)
		}
	}
	if explicit != "" {
		if err := ValidateID(explicit); err != nil {
			return "", fmt.Errorf("explicit identity: %w", err)
		}
		return explicit, nil
	}
	if session != "" {
		if err := ValidateID(session); err != nil {
			return "", fmt.Errorf("session identity: %w", err)
		}
		return session, nil
	}

	unique := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		unique[candidate] = struct{}{}
	}
	if len(unique) == 0 {
		return "", ErrIdentityMissing
	}
	if len(unique) == 1 {
		for candidate := range unique {
			return candidate, nil
		}
	}

	values := make([]string, 0, len(unique))
	for candidate := range unique {
		values = append(values, candidate)
	}
	sort.Strings(values)
	return "", fmt.Errorf("%w: candidates are %s", ErrIdentityAmbiguous, strings.Join(values, ", "))
}
