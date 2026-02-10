# Repositories

| Repo name                 | Purpose                | Source in this repo   |
| ------------------------- | ---------------------- | --------------------- |
| `golang-sdk-template`     | Go SDK starter         | `cmd/sdk/golang/`     |
| `python-sdk-template`     | Python SDK starter     | `cmd/sdk/python/`     |
| `typescript-sdk-template` | TypeScript SDK starter | `cmd/sdk/typescript/` |
| `powershell-sdk-template` | PowerShell starter     | `cmd/sdk/powershell/` |

## Creating a template repository

1. In GitHub, create a new repository under `sailpoint-oss` with the exact name (e.g. `golang-sdk-template`).
2. Clone it and copy the contents from the corresponding directory in this repo:
   - For **Go**: copy everything from `cmd/sdk/golang/`. Use `go.mod` as the filename (the source uses `go.mod.file` which is renamed when syncing to the template repo). Do not include `go.sum`; users run `go mod tidy` after init.
   - For **Python**, **TypeScript**, **PowerShell**: copy the contents of `cmd/sdk/python/`, `cmd/sdk/typescript/`, or `cmd/sdk/powershell/` as-is. Use `package.json` with a `{{.ProjectName}}` placeholder for the project name so the CLI can substitute it after fetch.
3. Ensure the default branch is `main`.
4. Push; the CLI will fetch from it when users run `sail sdk init <lang> [project-name]`.

## Template substitution

After fetching, the CLI applies Go `text/template` to any `package.json` or `connector-spec.json` in the extracted tree, with:

- `ProjectName`: the project directory name (e.g. `my-app`).

Template repos should use `{{.ProjectName}}` (or `{{$.ProjectName}}`) in those files where the project name is required.

## Requirements

- Users need network access to run `sail sdk init`; the CLI downloads the template archive from GitHub.
- If a template repo is missing or the request fails (e.g. 404, network error), the command returns an error. There is no offline or fallback mode.
