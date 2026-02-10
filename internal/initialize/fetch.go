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

	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.tar.gz", repoOwner, repoName, branch)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch template: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch template: HTTP %d", resp.StatusCode)
	}
	return ExtractAndInitProject(resp.Body, projName)
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

	gzr, err := gzip.NewReader(tarball)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read archive: %w", err)
		}

		name := filepath.FromSlash(hdr.Name)
		parts := strings.SplitN(name, string(filepath.Separator), 2)
		if len(parts) < 2 {
			continue
		}
		relPath := parts[1]
		if relPath == "" {
			continue
		}

		destPath := filepath.Join(projRoot, relPath)
		destPath = filepath.Clean(destPath)
		// Prevent Zip Slip / directory traversal: ensure destPath stays within projRoot.
		projRootWithSep := projRoot + string(os.PathSeparator)
		if destPath != projRoot && !strings.HasPrefix(destPath+string(os.PathSeparator), projRootWithSep) {
			return fmt.Errorf("unsafe path in archive entry %q", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
			}
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0755)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", destPath, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
			f.Close()
		}
	}

	if err := applyTemplatesInDir(projName, projName); err != nil {
		return err
	}
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
