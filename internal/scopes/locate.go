package scopes

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Source tells where the scope data for a language came from.
type Source struct {
	Language   Language `json:"language"`
	Embedded   bool     `json:"embedded"`
	Path       string   `json:"path,omitempty"`
	SDKVersion string   `json:"sdkVersion,omitempty"`
	Detail     string   `json:"detail,omitempty"`
}

const (
	manifestName = "sailpoint-scopes.json"
	goModulePath = "github.com/sailpoint-oss/golang-sdk/v3"
)

// locateManifest returns the path of the scope manifest shipped with the
// installed SDK for lang. dirs are the scan root, its parents and every
// project directory found during the walk. This search does not apply
// .gitignore, because node_modules and virtual environments are usually
// ignored.
func locateManifest(lang Language, dirs []string) (string, bool) {
	var candidates []string
	switch lang {
	case LangTypeScript:
		for _, d := range dirs {
			candidates = append(candidates, filepath.Join(d, "node_modules", "@sailpoint", "api-client", manifestName))
		}
	case LangAngular:
		for _, d := range dirs {
			candidates = append(candidates, filepath.Join(d, "node_modules", "@sailpoint", "angular-sdk", manifestName))
		}
	case LangPython:
		envs := []string{}
		for _, d := range dirs {
			for _, name := range []string{".venv", "venv", "env"} {
				envs = append(envs, filepath.Join(d, name))
			}
		}
		if v := os.Getenv("VIRTUAL_ENV"); v != "" {
			envs = append(envs, v)
		}
		for _, env := range envs {
			candidates = append(candidates, globAll(
				filepath.Join(env, "lib", "python*", "site-packages", "sailpoint", manifestName),
				filepath.Join(env, "Lib", "site-packages", "sailpoint", manifestName),
			)...)
		}
	case LangGo:
		for _, d := range dirs {
			candidates = append(candidates, filepath.Join(d, "vendor", filepath.FromSlash(goModulePath), manifestName))
		}
		if cache := goModCache(); cache != "" {
			for _, d := range dirs {
				if v := goModVersion(filepath.Join(d, "go.mod")); v != "" {
					candidates = append(candidates, filepath.Join(cache, filepath.FromSlash(goModulePath)+"@"+v, manifestName))
				}
			}
		}
	case LangPowerShell:
		for _, d := range psModulePaths() {
			matches := globAll(filepath.Join(d, "PSSailpoint", "*", manifestName))
			sort.Slice(matches, func(i, j int) bool {
				return versionLess(filepath.Base(filepath.Dir(matches[j])), filepath.Base(filepath.Dir(matches[i])))
			})
			candidates = append(candidates, matches...)
			candidates = append(candidates, filepath.Join(d, "PSSailpoint", manifestName))
		}
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, true
		}
	}
	return "", false
}

// searchDirs returns root, its parents and the project directories, without
// duplicates.
func searchDirs(root string, projectDirs []string) []string {
	seen := map[string]bool{}
	var dirs []string
	add := func(d string) {
		if !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	for d := root; ; d = filepath.Dir(d) {
		add(d)
		if filepath.Dir(d) == d {
			break
		}
	}
	for _, d := range projectDirs {
		add(d)
	}
	return dirs
}

func globAll(patterns ...string) []string {
	var out []string
	for _, p := range patterns {
		m, _ := filepath.Glob(p)
		out = append(out, m...)
	}
	return out
}

func goModCache() string {
	if c := os.Getenv("GOMODCACHE"); c != "" {
		return c
	}
	if p := os.Getenv("GOPATH"); p != "" {
		return filepath.Join(strings.Split(p, string(os.PathListSeparator))[0], "pkg", "mod")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "go", "pkg", "mod")
}

// goModVersion returns the golang-sdk version required by a go.mod file.
func goModVersion(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(sc.Text()), "require "))
		if len(fields) >= 2 && fields[0] == goModulePath {
			return fields[1]
		}
	}
	return ""
}

func psModulePaths() []string {
	if p := os.Getenv("PSModulePath"); p != "" {
		return strings.Split(p, string(os.PathListSeparator))
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return []string{filepath.Join(home, "Documents", "PowerShell", "Modules"), `C:\Program Files\PowerShell\Modules`}
	}
	return []string{filepath.Join(home, ".local", "share", "powershell", "Modules"), "/usr/local/share/powershell/Modules"}
}

// versionLess compares dotted numeric versions such as 1.6.10.
func versionLess(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			return x < y
		}
	}
	return false
}
