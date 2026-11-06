package scopes

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWalkRespectsGitignore(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".git/info/exclude":             "excluded.go\n",
		".gitignore":                    "build/\n*.gen.ts\n",
		"main.go":                       "package main",
		"excluded.go":                   "package main",
		"build/out.go":                  "package build",
		"src/app.ts":                    "",
		"src/app.gen.ts":                "",
		"src/.gitignore":                "local/\n!keep.gen.ts\n",
		"src/keep.gen.ts":               "",
		"src/local/skip.py":             "",
		"node_modules/@sailpoint/x.js":  "",
		"vendor/github.com/x/y.go":      "",
		"tools/package.json":            "{}",
		"tools/script.ps1":              "",
		"README.md":                     "",
		"nested/deep/dir/handler.py":    "",
		"nested/deep/dir/.gitignore":    "handler.py\n",
		"nested/deep/dir/kept_other.py": "",
	})

	res, err := walk(root)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range res.Files {
		got = append(got, f.Path)
	}
	sort.Strings(got)
	want := []string{
		"main.go",
		"nested/deep/dir/kept_other.py",
		"src/app.ts",
		"src/keep.gen.ts",
		"tools/script.ps1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
	if len(res.ProjectDirs) != 1 || filepath.Base(res.ProjectDirs[0]) != "tools" {
		t.Errorf("project dirs = %v, want [.../tools]", res.ProjectDirs)
	}
}

func TestWalkAppliesParentGitignore(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".git/HEAD":          "",
		".gitignore":         "generated/\n",
		"app/main.go":        "package main",
		"app/generated/x.go": "package generated",
	})

	res, err := walk(filepath.Join(root, "app"))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 1 || res.Files[0].Path != "main.go" {
		t.Errorf("files = %+v, want only main.go", res.Files)
	}
}
