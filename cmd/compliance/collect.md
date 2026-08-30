==Long==
# Collect

Collect compliance-relevant SailPoint tenant evidence and write it to a single JSON evidence bundle.

The command attempts all collectors, records failures in the summary, writes output, and returns a non-zero exit code when any collector fails.

====

==Example==
```bash
sail compliance collect --output evidence.json --period 90
sail compliance collect -o artifacts/evidence.json -p 30 --pretty
```
====
