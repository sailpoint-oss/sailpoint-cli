==Long==
# Update

Update a Workflow in IdentityNow

Arguments can be a list of directories or files. 
If a directory is specified, all JSON files in the directory will be parsed and the workflows uploaded.

## API References:
 - https://developer.sailpoint.com/idn/api/beta/update-workflow
====

==Example==
## File:
```bash
sail Workflow update -f {file-path} {file-path}
```

## Directory:
```bash
sail Workflow update -d {folder-path} {folder-path}
```
====