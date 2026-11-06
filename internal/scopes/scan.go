package scopes

import (
	"regexp"
	"strings"
)

// Call is one SDK call found in the scanned source.
type Call struct {
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Language    Language  `json:"language"`
	Name        string    `json:"name"`
	OperationID string    `json:"operationId"`
	Operation   Operation `json:"-"`
}

// extLanguages maps file extensions to the SDK languages they can use.
var extLanguages = map[string][]Language{
	".go":   {LangGo},
	".py":   {LangPython},
	".ts":   {LangTypeScript, LangAngular},
	".tsx":  {LangTypeScript, LangAngular},
	".js":   {LangTypeScript, LangAngular},
	".jsx":  {LangTypeScript, LangAngular},
	".mjs":  {LangTypeScript, LangAngular},
	".cjs":  {LangTypeScript, LangAngular},
	".ps1":  {LangPowerShell},
	".psm1": {LangPowerShell},
}

var (
	pythonImport = regexp.MustCompile(`(?m)^\s*(from|import)\s+sailpoint\b`)
	methodCall   = regexp.MustCompile(`\.\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	cmdletCall   = regexp.MustCompile(`\b([A-Za-z]+-[A-Za-z0-9]+)\b`)
)

// importsSDK reports whether a file references the SDK of lang. A project
// uses a language when at least one of its files does. PowerShell modules
// are often loaded by a profile, and cmdlet names are distinctive, so
// PowerShell files are always scanned.
func importsSDK(lang Language, content string) bool {
	switch lang {
	case LangGo:
		return strings.Contains(content, "github.com/sailpoint-oss/golang-sdk")
	case LangPython:
		return pythonImport.MatchString(content)
	case LangTypeScript:
		return strings.Contains(content, "@sailpoint/api-client")
	case LangAngular:
		return strings.Contains(content, "@sailpoint/angular-sdk")
	case LangPowerShell:
		return true
	}
	return false
}

// usedLanguages returns the SDK languages that the project imports.
func usedLanguages(files []sourceFile) map[Language]bool {
	used := map[Language]bool{}
	for _, f := range files {
		for _, lang := range extLanguages[f.Ext] {
			if !used[lang] && importsSDK(lang, f.Content) {
				used[lang] = true
			}
		}
	}
	return used
}

// fileLanguage picks the SDK language for one file. A JavaScript or
// TypeScript file is Angular when it imports the Angular SDK, or when the
// project uses only the Angular SDK.
func fileLanguage(f sourceFile, used map[Language]bool) (Language, bool) {
	langs := extLanguages[f.Ext]
	if len(langs) == 1 {
		return langs[0], used[langs[0]]
	}
	switch {
	case strings.Contains(f.Content, "@sailpoint/angular-sdk"):
		return LangAngular, true
	case strings.Contains(f.Content, "@sailpoint/api-client"):
		return LangTypeScript, true
	case used[LangTypeScript]:
		return LangTypeScript, true
	case used[LangAngular]:
		return LangAngular, true
	}
	return "", false
}

var pythonSuffixes = []string{"_with_http_info", "_without_preload_content"}

// scanFile finds the calls in f whose names are in index.
func scanFile(f sourceFile, lang Language, index map[string]string, ops map[string]Operation) []Call {
	var calls []Call
	for i, line := range strings.Split(f.Content, "\n") {
		if isCommentLine(lang, line) {
			continue
		}
		pattern := methodCall
		if lang == LangPowerShell {
			pattern = cmdletCall
		}
		for _, m := range pattern.FindAllStringSubmatch(line, -1) {
			name := m[1]
			key := name
			if lang == LangPowerShell {
				key = strings.ToLower(name)
			}
			id, ok := index[key]
			if !ok && lang == LangPython {
				for _, suffix := range pythonSuffixes {
					if base, cut := strings.CutSuffix(name, suffix); cut {
						id, ok = index[base]
						break
					}
				}
			}
			if !ok {
				continue
			}
			calls = append(calls, Call{
				File:        f.Path,
				Line:        i + 1,
				Language:    lang,
				Name:        name,
				OperationID: id,
				Operation:   ops[id],
			})
		}
	}
	return calls
}

// isCommentLine skips whole-line comments. Trailing comments and block
// comments that span lines without a leading marker are still scanned.
func isCommentLine(lang Language, line string) bool {
	t := strings.TrimSpace(line)
	switch lang {
	case LangPython, LangPowerShell:
		return strings.HasPrefix(t, "#")
	default:
		return strings.HasPrefix(t, "//") || strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "*")
	}
}
