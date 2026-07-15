package ui_plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

 // 100MB CLI guardrail; backend enforces the real limit
const maxBundleBytes = 100 << 20

// uploadFile is a single compiled asset to include in an asset bundle
// upload.
type uploadFile struct {
	// relPath is the bundle-relative path (forward slashes) sent
	// as the multipart filename; the backend preserves it as the
	// asset's path within the bundle.
	relPath string
	absPath string
}

// assetContentTypes maps a lowercase file extension to the single
// canonical MIME type the UMS asset-bundle endpoint accepts for that
// extension. The values are bare types (no charset parameter) so they
// exactly match the backend allowlist; mime.TypeByExtension is
// deliberately not used because it appends parameters
// (e.g. "text/javascript; charset=utf-8") that the backend rejects.
var assetContentTypes = map[string]string{
	".html":  "text/html",
	".htm":   "text/html",
	".js":    "text/javascript",
	".css":   "text/css",
	".json":  "application/json",
	".map":   "application/json",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".svg":   "image/svg+xml",
	".webp":  "image/webp",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".eot":   "application/vnd.ms-fontobject",
	".ico":   "image/x-icon",
	".txt":   "text/plain",
	".pdf":   "application/pdf",
}

// collectUploadFiles validates the build output directory and gathers
// the files to upload in a single pass. It fails fast (before any
// network call) when the directory is missing, is not a directory, or
// contains no uploadable files.  Dotfiles (e.g. .DS_Store) are
// skipped, matching the backend, which drops them.
func collectUploadFiles(root string) ([]uploadFile, error) {
	stat, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("output directory %s does not exist", root)
		}
		return nil, fmt.Errorf("cannot access output directory %s: %w", root, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("output path %s is not a directory", root)
	}

	var totalSize int64
	var errBundleTooLarge = errors.New("bundle exceeds upload guardrail")
	var files []uploadFile
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		totalSize += info.Size()
		if totalSize > maxBundleBytes {
			return errBundleTooLarge
		}
		files = append(files, uploadFile{relPath: filepath.ToSlash(rel), absPath: path})
		return nil
	})
	if errors.Is(err, errBundleTooLarge) {
		return nil, fmt.Errorf("output directory %s exceeds the %d MB upload guardrail", root, maxBundleBytes>>20)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory %s: %w", root, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("output directory %s contains no files to upload", root)
	}

	return files, nil
}

// contentTypeForPath returns the canonical MIME type for the file's
// extension, falling back to application/octet-stream for
// unrecognized extensions (which the backend rejects with an
// actionable message).
func contentTypeForPath(path string) string {
	if ct, ok := assetContentTypes[strings.ToLower(filepath.Ext(path))]; ok {
		return ct
	}
	return "application/octet-stream"
}

var multipartQuoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// buildAssetBundleBody encodes the files as a multipart/form-data
// body where each file is a part named "file" whose filename is the
// bundle-relative path and whose Content-Type matches the
// extension. It returns the body and the full content type (including
// the multipart boundary) to send as the request Content-Type.
func buildAssetBundleBody(files []uploadFile) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for _, f := range files {
		data, err := os.ReadFile(f.absPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read %s: %w", f.relPath, err)
		}

		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, multipartQuoteEscaper.Replace(f.relPath)))
		header.Set("Content-Type", contentTypeForPath(f.relPath))

		part, err := w.CreatePart(header)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", err
		}
	}

	if err := w.Close(); err != nil {
		return nil, "", err
	}

	return &buf, w.FormDataContentType(), nil
}

// assetBundleResponse is the subset of the UMS asset-bundle response used to
// render a success summary.
type assetBundleResponse struct {
	AssetBundleID string `json:"assetBundleId"`
	Assets        []struct {
		Path string `json:"path"`
	} `json:"assets"`
}

// renderUploadSuccess prints a human-readable confirmation of the
// uploaded bundle.
func renderUploadSuccess(w io.Writer, body []byte, alias string) error {
	var parsed assetBundleResponse
	_ = json.Unmarshal(body, &parsed)

	if parsed.AssetBundleID != "" {
		_, _ = fmt.Fprintf(w, "Uploaded %d asset(s) to plugin %q (bundle %s)\n", len(parsed.Assets), alias, parsed.AssetBundleID)
	} else {
		_, _ = fmt.Fprintf(w, "Uploaded assets to plugin %q\n", alias)
	}
	return nil
}

// mapUMSUploadError translates a non-2xx asset-bundle upload response
// into an actionable error, surfacing the message returned by the
// backend.
func mapUMSUploadError(status int, body []byte, alias string) error {
	message := umsErrorMessage(body)

	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid asset bundle for plugin %q: %s", alias, message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to upload UI plugin assets (requires the idn:plugins-ui:update right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("plugin instance for alias %q not found, or the UI plugins feature is not enabled for this tenant: %s", alias, message)
	case http.StatusRequestEntityTooLarge:
		return fmt.Errorf("asset bundle for plugin %q exceeds the maximum allowed size: %s", alias, message)
	default:
		return fmt.Errorf("failed to upload plugin assets (status %d): %s", status, message)
	}
}
