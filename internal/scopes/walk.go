package scopes

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// maxFileSize skips generated or minified files that are too large to be
// hand-written source.
const maxFileSize = 2 << 20

// skipDirs are never scanned, even when .gitignore does not list them. They
// hold dependencies, such as the SDK source itself, which would otherwise
// match every operation.
var skipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	".venv":         true,
	"venv":          true,
	"site-packages": true,
	"dist":          true,
	"__pycache__":   true,
}

// projectMarkers are files whose directory can hold an installed SDK, for
// example node_modules next to package.json.
var projectMarkers = map[string]bool{
	"go.mod":           true,
	"package.json":     true,
	"requirements.txt": true,
	"pyproject.toml":   true,
	"setup.py":         true,
}

// sourceFile is a file that the scanner reads.
type sourceFile struct {
	Path    string // relative to the scan root, with forward slashes
	Ext     string
	Content string
}

type walkResult struct {
	Files       []sourceFile
	ProjectDirs []string
}

// walk collects every source file under root that git would not ignore.
func walk(root string) (*walkResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	ig := newIgnoreSet(root)
	res := &walkResult{}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != root && (skipDirs[d.Name()] || ig.ignored(path, true)) {
				return filepath.SkipDir
			}
			ig.load(path)
			return nil
		}
		if !d.Type().IsRegular() || ig.ignored(path, false) {
			return nil
		}
		if projectMarkers[d.Name()] {
			res.ProjectDirs = append(res.ProjectDirs, filepath.Dir(path))
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := extLanguages[ext]; !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxFileSize {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		res.Files = append(res.Files, sourceFile{Path: filepath.ToSlash(rel), Ext: ext, Content: string(raw)})
		return nil
	})
	return res, err
}

// ignoreRule is one .gitignore line. Each line is compiled on its own so
// that a "!pattern" can re-include a path that a rule in a parent
// directory's .gitignore excluded. The library only applies negation within
// one file.
type ignoreRule struct {
	match  *ignore.GitIgnore
	negate bool
}

func readIgnoreRules(path string) []ignoreRule {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rules []ignoreRule
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		negate := strings.HasPrefix(line, "!")
		if negate {
			line = line[1:]
		}
		rules = append(rules, ignoreRule{match: ignore.CompileIgnoreLines(line), negate: negate})
	}
	return rules
}

// ignoreSet holds the .gitignore rules of every directory seen so far, keyed
// by absolute directory path.
type ignoreSet struct {
	base    string
	exclude []ignoreRule
	byDir   map[string][]ignoreRule
}

// newIgnoreSet loads the rules that apply above root: .git/info/exclude and
// the .gitignore files from the repository root down to root.
func newIgnoreSet(root string) *ignoreSet {
	s := &ignoreSet{base: root, byDir: map[string][]ignoreRule{}}
	gitRoot := findGitRoot(root)
	if gitRoot == "" {
		return s
	}
	s.base = gitRoot
	s.exclude = readIgnoreRules(filepath.Join(gitRoot, ".git", "info", "exclude"))
	for dir := gitRoot; ; {
		s.load(dir)
		if dir == root {
			break
		}
		rel, _ := filepath.Rel(dir, root)
		dir = filepath.Join(dir, strings.Split(rel, string(filepath.Separator))[0])
	}
	return s
}

func findGitRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func (s *ignoreSet) load(dir string) {
	if _, done := s.byDir[dir]; done {
		return
	}
	s.byDir[dir] = readIgnoreRules(filepath.Join(dir, ".gitignore"))
}

// ignored applies the rules from the repository root down to the parent of
// path. As in git, the last matching rule wins.
func (s *ignoreSet) ignored(path string, isDir bool) bool {
	rel, err := filepath.Rel(s.base, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	result := false
	apply := func(rules []ignoreRule, sub string) {
		if isDir {
			sub += "/"
		}
		for _, r := range rules {
			if r.match.MatchesPath(sub) {
				result = !r.negate
			}
		}
	}
	apply(s.exclude, strings.Join(parts, "/"))
	dir := s.base
	for i := range parts {
		apply(s.byDir[dir], strings.Join(parts[i:], "/"))
		dir = filepath.Join(dir, parts[i])
	}
	return result
}
