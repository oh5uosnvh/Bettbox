package main

import (
	"path/filepath"
	"testing"
)

func TestSharedRulesAcrossProfilesButNotCredentials(t *testing.T) {
	root := t.TempDir()
	makeRule := func(header string) map[string]map[string]any {
		return map[string]map[string]any{"r": {"type": "http", "url": "https://example.invalid/rules", "header": map[string]any{"Authorization": header}, "format": "yaml", "behavior": "domain"}}
	}
	a, b, c := makeRule("a"), makeRule("a"), makeRule("b")
	patchSharedRulePaths(a, filepath.Join(root, "one"))
	patchSharedRulePaths(b, filepath.Join(root, "two"))
	patchSharedRulePaths(c, filepath.Join(root, "three"))
	if a["r"]["path"] != b["r"]["path"] {
		t.Fatal("identical rules not shared")
	}
	if a["r"]["path"] == c["r"]["path"] {
		t.Fatal("credential scopes mixed")
	}
}
