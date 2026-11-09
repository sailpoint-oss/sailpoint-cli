# Scope snapshot

`sailpoint-scopes.json` is generated and is not committed. The CLI embeds it at
build time. To download the latest API spec and write the file, run this command
from the repository root:

```sh
go generate ./internal/scopes/...
```

You can also run `make generate`. If you build without the file, the build
succeeds, but scope commands that need the embedded snapshot return an error.
