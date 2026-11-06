// Command gen builds the scope snapshot embedded in the CLI from a
// dereferenced SailPoint OpenAPI bundle.
//
//	go run -tags scopesgen ./gen -spec deref-sailpoint-api.json -powershell-scopes PSSailpoint/sailpoint-scopes.json -out data/sailpoint-scopes.json
//
// Both inputs can be a file path or a URL. PowerShell cmdlet names come from
// the scope manifest that the PowerShell SDK publishes, because the SDK
// generator applies verb rules that the CLI does not copy. An operation that
// the PowerShell SDK does not have yet gets no PowerShell name.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/sailpoint-oss/sailpoint-cli/internal/scopes"
)

var httpMethods = []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

type specOperation struct {
	OperationID string                 `json:"operationId"`
	Tags        []string               `json:"tags"`
	Security    *[]map[string][]string `json:"security"`
	UserLevels  []string               `json:"x-sailpoint-userLevels"`
	// The misspelled key is present in a few spec files.
	UserLevelsTypo []string `json:"x-sailpont-userLevels"`
}

type spec struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
	Security []map[string][]string                 `json:"security"`
	Paths    map[string]map[string]json.RawMessage `json:"paths"`
}

func main() {
	specPath := flag.String("spec", "", "path or URL of a dereferenced OpenAPI JSON bundle")
	outPath := flag.String("out", "", "output file for the scope snapshot")
	psScopes := flag.String("powershell-scopes", "", "path or URL of the sailpoint-scopes.json published by the PowerShell SDK")
	flag.Parse()
	if *specPath == "" || *outPath == "" || *psScopes == "" {
		flag.Usage()
		os.Exit(2)
	}

	raw, err := readSource(*specPath)
	if err != nil {
		log.Fatalf("read spec: %v", err)
	}
	var s spec
	if err := json.Unmarshal(raw, &s); err != nil {
		log.Fatalf("parse spec: %v", err)
	}

	cmdlets, err := readCmdlets(*psScopes)
	if err != nil {
		log.Fatalf("read powershell scopes: %v", err)
	}

	m := scopes.Manifest{
		SchemaVersion: scopes.SchemaVersion,
		Source:        fmt.Sprintf("%s (version %s)", filepath.Base(*specPath), s.Info.Version),
		Operations:    map[string]scopes.Operation{},
	}

	missing := 0
	for path, item := range s.Paths {
		for _, method := range httpMethods {
			rawOp, ok := item[method]
			if !ok {
				continue
			}
			var so specOperation
			if err := json.Unmarshal(rawOp, &so); err != nil {
				log.Fatalf("parse %s %s: %v", method, path, err)
			}
			if so.OperationID == "" {
				continue
			}
			op := buildOperation(so, s.Security, strings.ToUpper(method), path)

			op.SDKNames = map[scopes.Language]string{
				scopes.LangGo:         goName(so.OperationID),
				scopes.LangPython:     pythonName(so.OperationID),
				scopes.LangTypeScript: so.OperationID,
			}
			if ps, ok := cmdlets[so.OperationID]; ok {
				op.SDKNames[scopes.LangPowerShell] = ps
			} else {
				missing++
			}
			m.Operations[so.OperationID] = op
		}
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Fatalf("encode: %v", err)
	}
	out = append(out, '\n')
	if err := os.MkdirAll(filepath.Dir(*outPath), 0755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}
	if err := os.WriteFile(*outPath, out, 0644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %d operations to %s (%d not in the PowerShell SDK)", len(m.Operations), *outPath, missing)
}

// readSource reads a local file, or downloads src when it is a URL.
func readSource(src string) ([]byte, error) {
	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		return os.ReadFile(src)
	}
	resp, err := http.Get(src)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", src, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// buildOperation applies the OpenAPI security rules. A missing security key
// inherits the global default. An empty list or an empty requirement object
// means that the operation is public.
func buildOperation(so specOperation, global []map[string][]string, method, path string) scopes.Operation {
	op := scopes.Operation{HTTPMethod: method, Path: path}
	if len(so.Tags) > 0 {
		op.Tag = so.Tags[0]
	}
	op.UserLevels = so.UserLevels
	if op.UserLevels == nil {
		op.UserLevels = so.UserLevelsTypo
	}

	security := global
	if so.Security != nil {
		security = *so.Security
	} else {
		op.DefaultScope = true
	}

	if len(security) == 0 {
		op.Public = true
		op.UserAuth = []string{}
		op.ApplicationAuth = []string{}
		return op
	}
	for _, req := range security {
		if len(req) == 0 {
			op.Public = true
			op.UserAuth = []string{}
			op.ApplicationAuth = []string{}
			return op
		}
		for scheme, list := range req {
			switch scheme {
			case "userAuth", "UserContextAuth":
				op.UserAuth = appendUnique(op.UserAuth, list)
			case "applicationAuth":
				op.ApplicationAuth = appendUnique(op.ApplicationAuth, list)
			}
		}
	}
	return op
}

func appendUnique(dst, src []string) []string {
	if dst == nil {
		dst = []string{}
	}
	for _, s := range src {
		found := false
		for _, d := range dst {
			if d == s {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, s)
		}
	}
	return dst
}

// goName matches the openapi-generator Go method name: the operationId with
// its first letter in upper case.
func goName(id string) string {
	r := []rune(id)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

var (
	underscoreFirst  = regexp.MustCompile(`([A-Z]+)([A-Z][a-z][a-z]+)`)
	underscoreSecond = regexp.MustCompile(`([a-z\d])([A-Z])`)
)

// pythonName matches the openapi-generator underscore() rule used for Python
// method names.
func pythonName(id string) string {
	s := underscoreFirst.ReplaceAllString(id, "${1}_${2}")
	s = underscoreSecond.ReplaceAllString(s, "${1}_${2}")
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ToLower(s)
}

// readCmdlets maps operationId to cmdlet name from the scope manifest that
// the PowerShell SDK publishes.
func readCmdlets(src string) (map[string]string, error) {
	raw, err := readSource(src)
	if err != nil {
		return nil, err
	}
	var ps struct {
		SDK        string `json:"sdk"`
		Operations map[string]struct {
			SDKName string `json:"sdkName"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(raw, &ps); err != nil {
		return nil, fmt.Errorf("parse %s: %w", src, err)
	}
	if ps.SDK != string(scopes.LangPowerShell) {
		return nil, fmt.Errorf("%s: sdk is %q, want %q", src, ps.SDK, scopes.LangPowerShell)
	}
	cmdlets := make(map[string]string, len(ps.Operations))
	for id, op := range ps.Operations {
		if op.SDKName != "" {
			cmdlets[id] = op.SDKName
		}
	}
	if len(cmdlets) == 0 {
		return nil, fmt.Errorf("%s: no cmdlet names found", src)
	}
	return cmdlets, nil
}
