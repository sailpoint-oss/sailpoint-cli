==Long==
# Parse

Parse log files from SailPoint VAs.
====

==Example==

## Parsing CCG Logs: 

All the errors will be parsed out of the log file and sorted by date and connector name.

Supplying the `--all` flag will parse all the log traffic out, not just errors.

```bash 
sail va parse --type ccg ./path/to/ccg.log ./path/to/ccg.log 
sail va parse --type ccg ./path/to/ccg.log ./path/to/ccg.log --all
```

## Parsing CANAL Logs: 

```bash
sail va parse --type canal ./path/to/canal.log ./path/to/canal.log 
```
====
