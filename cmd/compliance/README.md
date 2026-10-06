# Compliance

The `compliance` command provides evidence collection and control evaluation workflows for Identity Security Cloud tenants.

## Commands

### Collect evidence

```shell
sail compliance collect --output evidence.json --period 90 --pretty
```

This command gathers governance and security-relevant data into a single evidence bundle.

### Evaluate controls

```shell
sail compliance evaluate --input evidence.json --controls nist-800-53 --output findings.json
```

This command evaluates the evidence bundle against a control pack and emits findings.

You can also write a markdown report:

```shell
sail compliance evaluate --input evidence.json --output findings.json --output-md findings.md
```

## Control packs

The default embedded control pack is `nist-800-53`.

You can provide a custom control pack path:

```shell
sail compliance evaluate --input evidence.json --controls ./controls/custom.yaml
```

## Output schema

### Evidence bundle

The evidence bundle includes:

- `metadata`: schema and generation metadata
- `data`: raw API payloads by collector
- `summary`: collector success/failure summary

### Evaluation result

The evaluation result includes:

- `metadata`: copied from evidence bundle metadata
- `controls`: per-control and per-check status
- `findings`: failed check findings
- `summary`: roll-up counts including critical/high findings

## CI behavior

- `sail compliance collect` writes output even when some collectors fail, and returns non-zero if any collector fails.
- `sail compliance evaluate` writes outputs and returns non-zero when any checks fail.
