==Long==
# Evaluate

Evaluate an evidence bundle against a control pack and emit findings.

By default this command uses the embedded NIST 800-53 control pack. You can also pass a custom YAML control pack path.

The command writes output files first, then returns a non-zero exit code when controls fail.

====

==Example==
```bash
sail compliance evaluate --input evidence.json --controls nist-800-53
sail compliance evaluate -i evidence.json -c ./custom-controls.yaml -o findings.json --output-md findings.md
```
====
