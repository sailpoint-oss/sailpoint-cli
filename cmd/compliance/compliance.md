==Long==
# Compliance

Collect compliance evidence and evaluate against security control frameworks.

====

==Example==
```bash
sail compliance collect --output evidence.json --period 90
sail compliance evaluate --input evidence.json --controls nist-800-53
```
====
