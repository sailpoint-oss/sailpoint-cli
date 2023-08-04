==Long==
# Create
Create Workflows in IdentityNow

## API References:
 - https://developer.sailpoint.com/idn/api/beta/create-workflow

====

==Example==
## File Paths:
**Note:** File paths are relative to the current working directory, and only one workflow is allowed per file path. Multiple Workflows can be provided by specifying multiple file paths as arguments.

```bash
sail Workflow create -f {file-path}  
sail Workflow create -f {file-path} {file-path} ...
```

## Standard Input:
**Note:** Only one workflow is allowed via standard input.

```bash
sail Workflow create -s < {file-path}  
cat {file-path} | sail Workflow create -s
```
====