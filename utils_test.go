package main

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
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			t.Log(test.input)
			if got, expected := normalizeUri(test.input), test.want; got != expected {
				t.Fatalf("normalizeUri(%q) returned %q; expected %q", test.input, got, expected)
			}
		})
	}
}
