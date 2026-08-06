package workspace

import (
	"errors"
	"testing"
)

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "simple", value: "aiwf", wantErr: false},
		{name: "kebab case", value: "f0-workspace-information-architecture", wantErr: false},
		{name: "empty", value: "", wantErr: true},
		{name: "uppercase", value: "AIWF", wantErr: true},
		{name: "space", value: "aiwf core", wantErr: true},
		{name: "leading dash", value: "-aiwf", wantErr: true},
		{name: "trailing dash", value: "aiwf-", wantErr: true},
		{name: "double dash", value: "aiwf--core", wantErr: true},
		{name: "slash", value: "aiwf/core", wantErr: true},
		{name: "backslash", value: `aiwf\core`, wantErr: true},
		{name: "traversal", value: "..", wantErr: true},
		{name: "accented lowercase", value: "proyecto-á", wantErr: true},
		{name: "non-ascii digit", value: "aiwf-٣", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateID(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("error = %v, want ErrInvalidIdentity", err)
			}
		})
	}
}

func TestResolveIdentityPriority(t *testing.T) {
	tests := []struct {
		name       string
		explicit   string
		session    string
		candidates []string
		want       string
	}{
		{name: "explicit wins", explicit: "explicit", session: "session", candidates: []string{"only"}, want: "explicit"},
		{name: "session wins", session: "session", candidates: []string{"only"}, want: "session"},
		{name: "unique candidate", candidates: []string{"only"}, want: "only"},
		{name: "duplicate candidates collapse", candidates: []string{"only", "only"}, want: "only"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveIdentity(tt.explicit, tt.session, tt.candidates)
			if err != nil {
				t.Fatalf("ResolveIdentity: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ResolveIdentity = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveIdentityErrors(t *testing.T) {
	tests := []struct {
		name       string
		explicit   string
		session    string
		candidates []string
		want       error
	}{
		{name: "missing", want: ErrIdentityMissing},
		{name: "ambiguous", candidates: []string{"beta", "alpha"}, want: ErrIdentityAmbiguous},
		{name: "invalid explicit", explicit: "../bad", want: ErrInvalidIdentity},
		{name: "invalid session", session: "Bad", want: ErrInvalidIdentity},
		{name: "invalid candidate", candidates: []string{"valid", "bad/path"}, want: ErrInvalidIdentity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveIdentity(tt.explicit, tt.session, tt.candidates)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}
