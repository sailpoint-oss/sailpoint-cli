==Long==

# Reassign

Run this command to reassign object ownership in Identity Security Cloud.

You can reassign ownership for the following supported object types:

* sources
* roles
* access profiles
* entitlements
* identity profiles
* governance groups
* workflows

====

==Example==

Reassign all objects from one identity to another

```bash
sail reassign --from <old-owner-id> --to <new-owner-id> 
```

Reassign a subset of supported object types

```bash
sail reassign --from <old-owner-id> --to <new-owner-id> --object-types role,access-profile
```

Reassign a single object

```bash
sail reassign --from <old-owner-id> --to <new-owner-id> --object-id <object-id>
```

Reassign a single object without confirmation

```bash
sail reassign --from <old-owner-id> --to <new-owner-id> --object-id <object-id> --force
```

Use the `--dry-run` flag to preview which objects will be reassigned, without making any changes.

```bash
sail reassign --from <old-owner-id> --to <new-owner-id> --dry-run
```

====
