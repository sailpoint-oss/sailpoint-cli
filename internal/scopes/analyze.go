package scopes

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Options configures Analyze.
type Options struct {
	Root string
	Auth AuthType
	// Manifests overrides the scope data for a language. Tests use it to
	// avoid the installed SDKs and the embedded snapshot.
	Manifests map[Language]*Manifest
	// Locate finds an installed SDK manifest. It defaults to locateManifest.
	Locate func(lang Language, dirs []string) (string, bool)
}

// Analyze scans opts.Root for SailPoint SDK calls and returns the scopes that
// they need.
func Analyze(opts Options) (*Result, error) {
	if opts.Auth == "" {
		opts.Auth = AuthUser
	}
	if opts.Locate == nil {
		opts.Locate = locateManifest
	}
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	walked, err := walk(root)
	if err != nil {
		return nil, err
	}
	used := usedLanguages(walked.Files)
	dirs := searchDirs(root, walked.ProjectDirs)

	type langData struct {
		manifest *Manifest
		index    map[string]string
		source   Source
	}
	data := map[Language]*langData{}
	var snap *Manifest
	load := func(lang Language) (*langData, error) {
		if d, ok := data[lang]; ok {
			return d, nil
		}
		d := &langData{source: Source{Language: lang}}
		if m, ok := opts.Manifests[lang]; ok {
			d.manifest = m
			d.source.SDKVersion = m.SDKVersion
			d.source.Detail = "provided"
		} else if path, ok := opts.Locate(lang, dirs); ok {
			m, err := Load(path)
			if err != nil {
				return nil, err
			}
			d.manifest = m
			d.source.Path = path
			d.source.SDKVersion = m.SDKVersion
		} else {
			if snap == nil {
				if snap, err = Snapshot(); err != nil {
					return nil, err
				}
			}
			d.manifest = snap
			d.source.Embedded = true
			d.source.Detail = snap.Source
		}
		d.index = d.manifest.NameIndex(lang)
		data[lang] = d
		return d, nil
	}

	res := &Result{AuthType: opts.Auth, Calls: []Call{}, Warnings: []string{}}
	callLangs := map[Language]bool{}
	for _, f := range walked.Files {
		lang, ok := fileLanguage(f, used)
		if !ok {
			continue
		}
		d, err := load(lang)
		if err != nil {
			return nil, err
		}
		calls := scanFile(f, lang, d.index, d.manifest.Operations)
		if len(calls) > 0 {
			callLangs[lang] = true
			res.Calls = append(res.Calls, calls...)
		}
	}
	sort.SliceStable(res.Calls, func(i, j int) bool {
		if res.Calls[i].File != res.Calls[j].File {
			return res.Calls[i].File < res.Calls[j].File
		}
		return res.Calls[i].Line < res.Calls[j].Line
	})

	res.Sources = []Source{}
	for _, lang := range Languages {
		if !callLangs[lang] {
			continue
		}
		src := data[lang].source
		res.Sources = append(res.Sources, src)
		if src.Embedded {
			res.Warnings = append(res.Warnings, fmt.Sprintf(
				"no installed %s SDK scope manifest found; using the scopes embedded in this CLI, which can differ from your SDK version", lang))
		}
	}

	var warnings []string
	res.Operations, res.RequiredScopes, warnings = Resolve(res.Calls, opts.Auth)
	res.Warnings = append(res.Warnings, warnings...)
	if res.RequiredScopes == nil {
		res.RequiredScopes = []string{}
	}
	return res, nil
}
