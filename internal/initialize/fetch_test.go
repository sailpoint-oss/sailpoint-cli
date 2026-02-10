// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package initialize

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	goTarballRoot     = "golang-sdk-template-main"
	tsTarballRoot     = "typescript-sdk-template-main"
	pythonTarballRoot = "python-sdk-template-main"
)

func buildGolangTarball() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	root := goTarballRoot + "/"
	_ = tw.WriteHeader(&tar.Header{Name: root, Typeflag: tar.TypeDir, Mode: 0755})
	goMod := "module example.com/app\n\ngo 1.21\n"
	_ = tw.WriteHeader(&tar.Header{Name: root + "go.mod", Typeflag: tar.TypeReg, Size: int64(len(goMod)), Mode: 0644})
	_, _ = tw.Write([]byte(goMod))
	mainGo := "package main\n\nfunc main() {}\n"
	_ = tw.WriteHeader(&tar.Header{Name: root + "main.go", Typeflag: tar.TypeReg, Size: int64(len(mainGo)), Mode: 0644})
	_, _ = tw.Write([]byte(mainGo))
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

func buildTypeScriptTarball() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	root := tsTarballRoot + "/"
	_ = tw.WriteHeader(&tar.Header{Name: root, Typeflag: tar.TypeDir, Mode: 0755})
	_ = tw.WriteHeader(&tar.Header{Name: root + "src/", Typeflag: tar.TypeDir, Mode: 0755})
	pkgJSON := `{"name":"{{.ProjectName}}","version":"1.0.0","main":"build/index.js","types":"build/index.d.ts","scripts":{"build":"tsc"},"devDependencies":{"typescript":"^5.0.0"}}`
	_ = tw.WriteHeader(&tar.Header{Name: root + "package.json", Typeflag: tar.TypeReg, Size: int64(len(pkgJSON)), Mode: 0644})
	_, _ = tw.Write([]byte(pkgJSON))
	tsconfig := `{"compilerOptions":{"target":"ES2020","module":"commonjs","outDir":"./build","rootDir":"./src"},"include":["src/**/*"]}`
	_ = tw.WriteHeader(&tar.Header{Name: root + "tsconfig.json", Typeflag: tar.TypeReg, Size: int64(len(tsconfig)), Mode: 0644})
	_, _ = tw.Write([]byte(tsconfig))
	indexTs := "const x = 1;\nexport {};\n"
	_ = tw.WriteHeader(&tar.Header{Name: root + "src/index.ts", Typeflag: tar.TypeReg, Size: int64(len(indexTs)), Mode: 0644})
	_, _ = tw.Write([]byte(indexTs))
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

func buildPythonTarball() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	root := pythonTarballRoot + "/"
	_ = tw.WriteHeader(&tar.Header{Name: root, Typeflag: tar.TypeDir, Mode: 0755})
	reqTxt := "# SDK placeholder\n"
	_ = tw.WriteHeader(&tar.Header{Name: root + "requirements.txt", Typeflag: tar.TypeReg, Size: int64(len(reqTxt)), Mode: 0644})
	_, _ = tw.Write([]byte(reqTxt))
	sdkPy := `print("ok")
`
	_ = tw.WriteHeader(&tar.Header{Name: root + "sdk.py", Typeflag: tar.TypeReg, Size: int64(len(sdkPy)), Mode: 0644})
	_, _ = tw.Write([]byte(sdkPy))
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

func TestFetchAndInitProject_EmptyProjectName(t *testing.T) {
	err := FetchAndInitProject("org", "repo", "", "")
	if err == nil {
		t.Fatal("expected error for empty project name")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("error should mention empty, got: %v", err)
	}
}

func TestFetchAndInitProject_EmptyRepoOwner(t *testing.T) {
	err := FetchAndInitProject("", "repo", "", "myapp")
	if err == nil {
		t.Fatal("expected error for empty owner")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error should mention required, got: %v", err)
	}
}

func TestFetchAndInitProject_ProjectAlreadyExists(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	projName := "exists"
	if err := os.Mkdir(projName, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	err = FetchAndInitProject("org", "repo", "", projName)
	if err == nil {
		t.Fatal("expected error when project already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention already exists, got: %v", err)
	}
}

func TestExtractAndInitProject_ProjectAlreadyExists(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	projName := "exists"
	if err := os.Mkdir(projName, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	err = ExtractAndInitProject(bytes.NewReader(buildGolangTarball()), projName)
	if err == nil {
		t.Fatal("expected error when project already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention already exists, got: %v", err)
	}
}

func TestExtractAndInitProject_EmptyProjectName(t *testing.T) {
	err := ExtractAndInitProject(bytes.NewReader(buildGolangTarball()), "")
	if err == nil {
		t.Fatal("expected error for empty project name")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("error should mention empty, got: %v", err)
	}
}

func TestExtractAndInitProject_Go_Builds(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	projName := "my-go-app"
	err = ExtractAndInitProject(bytes.NewReader(buildGolangTarball()), projName)
	if err != nil {
		t.Fatalf("ExtractAndInitProject: %v", err)
	}

	projectPath := filepath.Join(workdir, projName)
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err != nil {
		t.Fatalf("go.mod missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectPath, "main.go")); err != nil {
		t.Fatalf("main.go missing: %v", err)
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
}

func TestExtractAndInitProject_TypeScript_Builds(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	projName := "my-ts-app"
	err = ExtractAndInitProject(bytes.NewReader(buildTypeScriptTarball()), projName)
	if err != nil {
		t.Fatalf("ExtractAndInitProject: %v", err)
	}

	projectPath := filepath.Join(workdir, projName)
	pkgPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	if !strings.Contains(string(data), projName) {
		t.Errorf("package.json should contain project name %q, got %s", projName, data)
	}

	cmd := exec.Command("npm", "install")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm install failed: %v\n%s", err, out)
	}
	cmd = exec.Command("npm", "run", "build")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm run build failed: %v\n%s", err, out)
	}
}

func TestExtractAndInitProject_Python_Compiles(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	projName := "my-py-app"
	err = ExtractAndInitProject(bytes.NewReader(buildPythonTarball()), projName)
	if err != nil {
		t.Fatalf("ExtractAndInitProject: %v", err)
	}

	projectPath := filepath.Join(workdir, projName)
	sdkPath := filepath.Join(projectPath, "sdk.py")
	if _, err := os.Stat(sdkPath); err != nil {
		t.Fatalf("sdk.py missing: %v", err)
	}

	cmd := exec.Command("python3", "-m", "py_compile", "sdk.py")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("python3 -m py_compile sdk.py failed: %v\n%s", err, out)
	}
}
