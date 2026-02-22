package util //nolint:testpackage,revive // white-box testing of internal package

import (
	"testing"
)

func TestNormalizeUri(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{input: "redis://password@host/db?query=param", want: "redis://:password@host/db?query=param"},
		{input: "redis://:password@host/db?query=param", want: "redis://:password@host/db?query=param"},
		{input: "rediss://host:1234/db?query=param", want: "rediss://host:1234/db?query=param"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			t.Log(test.input)
			got, err := NormalizeURI(test.input)
			if err != nil {
				t.Fatalf("normalizeUri(%q) returned unexpected error: %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("normalizeUri(%q) returned %q; expected %q", test.input, got, test.want)
			}
		})
	}
}
