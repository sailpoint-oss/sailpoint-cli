// Command gen builds the scope snapshot embedded in the CLI from a
// dereferenced SailPoint OpenAPI bundle.
//
//	go run ./gen -spec deref-sailpoint-api.json -out data/sailpoint-scopes.json [-powershell-sdk DIR]
//
// When -powershell-sdk points at a powershell-sdk checkout, cmdlet names are
// read from the generated Api/*.ps1 files. Otherwise they are derived from the
// operationId with the same verb rules as the SDK generator.
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
	"sort"
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
	psSDK := flag.String("powershell-sdk", "", "optional powershell-sdk checkout to read exact cmdlet names from")
	flag.Parse()
	if *specPath == "" || *outPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	raw, err := readSpec(*specPath)
	if err != nil {
		log.Fatalf("read spec: %v", err)
	}
	var s spec
	if err := json.Unmarshal(raw, &s); err != nil {
		log.Fatalf("parse spec: %v", err)
	}

	var cmdlets map[string]string
	if *psSDK != "" {
		cmdlets, err = readCmdlets(*psSDK)
		if err != nil {
			log.Fatalf("read powershell sdk: %v", err)
		}
	}

	m := scopes.Manifest{
		SchemaVersion: scopes.SchemaVersion,
		Source:        fmt.Sprintf("%s (version %s)", filepath.Base(*specPath), s.Info.Version),
		Operations:    map[string]scopes.Operation{},
	}

	derived := 0
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

			ps, ok := cmdlets[strings.ToUpper(method)+" "+path]
			if !ok {
				ps = powershellName(so.OperationID)
				derived++
			}
			op.SDKNames = map[scopes.Language]string{
				scopes.LangGo:         goName(so.OperationID),
				scopes.LangPython:     pythonName(so.OperationID),
				scopes.LangTypeScript: so.OperationID,
				scopes.LangPowerShell: ps,
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
	log.Printf("wrote %d operations to %s (%d derived cmdlet names)", len(m.Operations), *outPath, derived)
}

func readSpec(src string) ([]byte, error) {
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

// commonVerbs maps operationId verbs that PowerShell does not approve to the
// verb the SDK generator uses instead.
var commonVerbs = map[string]string{
	"Create": "New", "Delete": "Remove", "Cancel": "Suspend", "Patch": "Update",
	"List": "Get", "Query": "Search", "Run": "Start", "Validate": "Confirm", "Make": "Select",
	"Post": "New", "Put": "Send", "Generate": "New", "Reject": "Deny",
}

var approvedVerbs = []string{
	"Add", "Approve", "Clear", "Close", "Compare", "Complete", "Copy", "Deny", "Disable",
	"Enable", "Export", "Get", "Import", "Move", "Ping", "Publish", "Remove", "Reset",
	"Search", "Select", "Send", "Set", "Show", "Start", "Submit", "Suspend", "Sync",
	"Test", "Unlock", "Update",
}

// powershellName derives a cmdlet name the way the SDK generator does. It is
// only used when no powershell-sdk checkout is given.
func powershellName(id string) string {
	name := goName(id)
	for _, verb := range approvedVerbs {
		if strings.HasPrefix(name, verb) {
			return verb + "-" + name[len(verb):]
		}
	}
	for verb, replacement := range commonVerbs {
		if strings.HasPrefix(name, verb) {
			return replacement + "-" + name[len(verb):]
		}
	}
	return "Invoke-" + name
}

var (
	psFunction = regexp.MustCompile(`(?m)^function ([A-Za-z]+-[A-Za-z0-9]+)\s*\{`)
	psURI      = regexp.MustCompile(`\$LocalVarUri = '([^']+)'`)
	psMethod   = regexp.MustCompile(`Invoke-ApiClient -Method '([A-Z]+)'`)
)

// readCmdlets maps "METHOD /path" to the cmdlet name for every generated
// PowerShell API function in the checkout.
func readCmdlets(root string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(root, "PSSailpoint", "*", "src", "*", "Api", "*.ps1"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no PSSailpoint/*/src/*/Api/*.ps1 files under %s", root)
	}
	sort.Strings(files)
	cmdlets := map[string]string{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		text := string(raw)
		locs := psFunction.FindAllStringSubmatchIndex(text, -1)
		for i, loc := range locs {
			end := len(text)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			body := text[loc[1]:end]
			uri := psURI.FindStringSubmatch(body)
			method := psMethod.FindStringSubmatch(body)
			if uri == nil || method == nil {
				continue
			}
			cmdlets[method[1]+" "+uri[1]] = text[loc[2]:loc[3]]
		}
	}
	return cmdlets, nil
}
