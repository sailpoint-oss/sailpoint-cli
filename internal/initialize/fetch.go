// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package initialize

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
)

const (
	defaultBranch = "main"
)

// FetchAndInitProject downloads a GitHub repository archive, extracts it into
// projName, and applies template substitutions (e.g. ProjectName in package.json).
// repoOwner and repoName are the GitHub org/repo (e.g. "sailpoint-oss", "golang-sdk-template").
// branch is the git branch or tag to fetch; if empty, "main" is used.
func FetchAndInitProject(repoOwner, repoName, branch, projName string) error {
	if projName == "" {
		return errors.New("project name cannot be empty")
	}
	if repoOwner == "" || repoName == "" {
		return errors.New("repo owner and name are required")
	}
	if branch == "" {
		branch = defaultBranch
	}
	if f, err := os.Stat(projName); err == nil && f.IsDir() && f.Name() == projName {
		return fmt.Errorf("error: project '%s' already exists", projName)
	}

	body, err := FetchArchive(repoOwner, repoName, branch)
	if err != nil {
		return err
	}
	defer body.Close()
	log.Info("Download complete. Extracting and initializing project...", "project", projName)
	return ExtractAndInitProject(body, projName)
}

// ExtractAndInitProject extracts a gzipped tar archive from tarball into projName
// (stripping the archive root directory) and applies template substitutions.
// Used by FetchAndInitProject and by tests with testdata tarballs.
func ExtractAndInitProject(tarball io.Reader, projName string) error {
	if projName == "" {
		return errors.New("project name cannot be empty")
	}
	if f, err := os.Stat(projName); err == nil && f.IsDir() && f.Name() == projName {
		return fmt.Errorf("error: project '%s' already exists", projName)
	}

	projRoot, err := filepath.Abs(projName)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	log.Info("Extracting archive", "project", projName)
	if _, err := extractGzipTar(tarball, projRoot, stripArchiveRoot); err != nil {
		return err
	}

	log.Info("Applying template substitutions...")
	if err := applyTemplatesInDir(projName, projName); err != nil {
		return err
	}
	log.Info("Project created successfully.", "project", projName)
	printDir(projName, 0)
	return nil
}

// applyTemplatesInDir walks dir and applies Go template substitution to
// package.json and connector-spec.json using ProjectName = projName.
func applyTemplatesInDir(dir, projName string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if err := applyTemplatesInDir(full, projName); err != nil {
				return err
			}
			continue
		}
		switch e.Name() {
		case "package.json", "connector-spec.json":
			if err := applyTemplateFile(full, projName); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyTemplateFile(filePath, projName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	t, err := template.New(filepath.Base(filePath)).Parse(string(data))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	templateData := struct {
		ProjectName string
	}{ProjectName: projName}
	if err := t.Execute(&buf, templateData); err != nil {
		return err
	}
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// FetchArchive downloads a GitHub repository branch archive (.tar.gz) and returns
// the response body for streaming extraction. The caller must Close it. branch
// defaults to "main" when empty. This is the download half of
// FetchAndInitProject, factored out so callers can extract a subtree
// (ExtractSubtree) or otherwise handle the archive themselves.
func FetchArchive(repoOwner, repoName, branch string) (io.ReadCloser, error) {
	if repoOwner == "" || repoName == "" {
		return nil, errors.New("repo owner and name are required")
	}
	if branch == "" {
		branch = defaultBranch
	}
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.tar.gz", repoOwner, repoName, branch)
	log.Info("Fetching template from GitHub", "owner", repoOwner, "repo", repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch template: HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// ExtractSubtree extracts only the entries under <archiveRoot>/<subdir>/ from a
// gzipped tar archive into destDir, stripping both the archive root and the
// subdir prefix. Unlike ExtractAndInitProject it applies no template
// substitution: templates that must stay runnable and browsable are
// personalized by the caller after extraction. It fails if destDir already
// exists or if the requested subdir is absent from the archive.
func ExtractSubtree(tarball io.Reader, destDir, subdir string) error {
	if destDir == "" {
		return errors.New("destination directory cannot be empty")
	}
	if strings.TrimSpace(subdir) == "" {
		return errors.New("subdir cannot be empty")
	}
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("error: project '%s' already exists", destDir)
	}

	destRoot, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Compare in OS-separator space to match extractGzipTar's entry handling.
	osSubdir := filepath.Clean(filepath.FromSlash(strings.Trim(subdir, "/")))
	prefix := osSubdir + string(filepath.Separator)

	// strip keeps only entries under the archive root's <subdir>/ and returns
	// their path relative to that subdir. The subdir's own directory entry maps
	// to an empty relPath, which extractGzipTar skips.
	strip := func(osName string) (string, bool) {
		parts := strings.SplitN(osName, string(filepath.Separator), 2)
		if len(parts) < 2 {
			return "", false
		}
		belowRoot := parts[1]
		if belowRoot == osSubdir {
			return "", true
		}
		if !strings.HasPrefix(belowRoot, prefix) {
			return "", false
		}
		return strings.TrimPrefix(belowRoot, prefix), true
	}

	log.Info("Extracting subtree", "subdir", subdir, "dest", destDir)
	written, err := extractGzipTar(tarball, destRoot, strip)
	if err != nil {
		return err
	}
	if written == 0 {
		return fmt.Errorf("subdir %q not found in archive", subdir)
	}
	log.Info("Subtree extracted.", "dest", destDir)
	return nil
}

// stripArchiveRoot removes the single top-level directory component that GitHub
// archives wrap everything in (e.g. "repo-main/"). It is the default strip used
// by ExtractAndInitProject.
func stripArchiveRoot(osName string) (string, bool) {
	parts := strings.SplitN(osName, string(filepath.Separator), 2)
	if len(parts) < 2 {
		return "", false
	}
	return parts[1], true
}

// extractGzipTar streams a gzipped tar archive, applies strip() to each entry to
// compute its destination-relative path (dropping entries strip rejects or maps
// to ""), enforces path-traversal / Zip-Slip safety, and writes files and
// directories under destRoot. It returns the number of entries written so
// callers can detect an empty/absent selection.
func extractGzipTar(tarball io.Reader, destRoot string, strip func(osName string) (relPath string, keep bool)) (int, error) {
	gzr, err := gzip.NewReader(tarball)
	if err != nil {
		return 0, fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzr.Close()

	written := 0
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return written, fmt.Errorf("failed to read archive: %w", err)
		}

		name := filepath.FromSlash(hdr.Name)
		relPath, keep := strip(name)
		if !keep || relPath == "" {
			continue
		}
		// Reject path traversal (".." components) and absolute paths in archive entry names.
		if filepath.IsAbs(relPath) {
			return written, fmt.Errorf("unsafe path in archive entry %q", hdr.Name)
		}
		for _, part := range strings.Split(relPath, string(filepath.Separator)) {
			if part == ".." {
				return written, fmt.Errorf("unsafe path in archive entry %q", hdr.Name)
			}
		}

		destPath := filepath.Clean(filepath.Join(destRoot, relPath))
		// Prevent Zip Slip: ensure resolved path stays under destRoot (filepath.Rel is the standard check).
		relToRoot, relErr := filepath.Rel(destRoot, destPath)
		if relErr != nil {
			return written, fmt.Errorf("unsafe path in archive entry %q: %w", hdr.Name, relErr)
		}
		if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(os.PathSeparator)) {
			return written, fmt.Errorf("unsafe path in archive entry %q", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return written, fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			written++
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return written, fmt.Errorf("failed to create directory for %s: %w", destPath, err)
			}
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0755)
			if err != nil {
				return written, fmt.Errorf("failed to create file %s: %w", destPath, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return written, fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
			f.Close()
			written++
		}
	}
	return written, nil
}
