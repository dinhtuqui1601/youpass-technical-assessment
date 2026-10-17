package share

import (
	"strings"
	"testing"
)

func TestGenerateCodeLengthAndAlphabet(t *testing.T) {
	for _, length := range []int{1, 4, 10, 16} {
		code, err := GenerateCode(length)
		if err != nil {
			t.Fatalf("GenerateCode(%d) error: %v", length, err)
		}
		if len(code) != length {
			t.Fatalf("GenerateCode(%d) length = %d, want %d", length, len(code), length)
		}
		for _, r := range code {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("GenerateCode produced invalid character %q", r)
			}
		}
	}
}

func TestGenerateCodeUniqueness(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 5000; i++ {
		code, err := GenerateCode(10)
		if err != nil {
			t.Fatalf("GenerateCode error: %v", err)
		}
		if _, ok := seen[code]; ok {
			t.Fatalf("duplicate code generated: %s", code)
		}
		seen[code] = struct{}{}
	}
}

func TestGenerateCodeInvalidLength(t *testing.T) {
	for _, length := range []int{0, -1} {
		if _, err := GenerateCode(length); err == nil {
			t.Fatalf("GenerateCode(%d) expected error, got nil", length)
		}
	}
}

func TestValidCodeFormat(t *testing.T) {
	cases := []struct {
		code string
		want bool
	}{
		{"aB7xK2pQ9z", true},
		{"", false},
		{"has space", false},
		{"with-dash", false},
		{"a!b", false},
		{strings.Repeat("a", 65), false},
	}
	for _, tc := range cases {
		if got := ValidCodeFormat(tc.code); got != tc.want {
			t.Errorf("ValidCodeFormat(%q) = %v, want %v", tc.code, got, tc.want)
		}
	}
}
