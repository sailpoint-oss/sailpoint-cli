==Long==
# Create
Create workflows in IdentityNow. 

## API References:
 - https://developer.sailpoint.com/idn/api/beta/create-workflow

====

==Example==
## File Paths:
**Note:** File paths are relative to the current working directory, and only one workflow is allowed per file path. Multiple workflows can be provided by specifying multiple file paths as arguments.

```bash
sail workflow create -f {file-path}  
sail workflow create -f {file-path} {file-path} ...
```

## Folder Paths:
```bash
sail workflow create -d {folder-path}
sail workflow create -d {folder-path} {folder-path} ...
```

====