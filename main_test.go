package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args     []string
		tagger   tagger
		expected vtag
	}{
		"major": {
			args: []string{"/program", "major"},
			tagger: &memoryTagger{
				all: []vtag{
					{major: 0, minor: 1, patch: 3},
					{major: 0, minor: 2, patch: 1},
					{major: 1, minor: 2, patch: 3},
				},
			},
			expected: vtag{major: 2, minor: 0, patch: 0},
		},
		"minor": {
			args: []string{"/program", "minor"},
			tagger: &memoryTagger{
				all: []vtag{
					{major: 0, minor: 1, patch: 3},
					{major: 0, minor: 2, patch: 1},
					{major: 1, minor: 2, patch: 3},
				},
			},
			expected: vtag{major: 1, minor: 3, patch: 0},
		},
		"patch": {
			args: []string{"/program", "patch"},
			tagger: &memoryTagger{
				all: []vtag{
					{major: 0, minor: 1, patch: 3},
					{major: 0, minor: 2, patch: 1},
					{major: 1, minor: 2, patch: 3},
				},
			},
			expected: vtag{major: 1, minor: 2, patch: 4},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := run(test.tagger, test.args); err != nil {
				t.Errorf("should have run: %w", err)
			}

			actual, _ := test.tagger.latest()
			assertVtag(t, actual, test.expected)
		})
	}
}

func TestVtag_Scan(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    string
		expected vtag
	}{
		"v0.0.0": {
			input: "v0.0.0",
			expected: vtag{
				major: 0, minor: 0, patch: 0,
			},
		},
		"v1.2.3": {
			input: "v1.2.3",
			expected: vtag{
				major: 1, minor: 2, patch: 3,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual vtag
			if _, err := fmt.Fscan(strings.NewReader(test.input), &actual); err != nil {
				t.Errorf("should have scan the input: %w", err)
			}

			assertVtag(t, actual, test.expected)
		})
	}
}

func assertVtag(t *testing.T, actual, expected vtag) {
	assertEqual(t, "major", actual.major, expected.major)
	assertEqual(t, "minor", actual.minor, expected.minor)
	assertEqual(t, "patch", actual.patch, expected.patch)
}

func assertEqual(t *testing.T, name string, actual, expected interface{}) {
	if actual != expected {
		t.Errorf("%s: expected %v to equal to %v", name, actual, expected)
	}
}

type memoryTagger struct {
	all []vtag
}

func (t memoryTagger) tags() ([]vtag, error) {
	sortVtags(t.all)
	return t.all, nil
}

func (t memoryTagger) latest() (vtag, error) {
	return latestVtag(t.all), nil
}

func (t *memoryTagger) create(tag vtag) error {
	t.all = append(t.all, tag)
	return nil
}
