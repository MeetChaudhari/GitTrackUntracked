# Architecture

`gitu` is a small Go CLI that orchestrates ordinary Git repositories rather
than replacing Git.

```text
host checkout                     personal vault (one remote)
-------------                     ---------------------------
tracked source files ── untouched  projects/<project-id>/...
untracked, registered paths ─────► files/<path>
.git/gittrackuntracked/config.json   manifest.json + project.json
```

## Local project configuration

The configuration sits under the host repository's `.git` directory. It stores
the project ID, storage mode, and explicitly registered relative paths. Because
it is outside the work tree, it cannot appear in normal Git status or commits.

## Vault data

The vault is cloned under the platform user configuration directory, or the
directory selected by `GITU_VAULT`. It has an `origin` remote set by the user.
All changes are normal Git commits and are pushed only by `gitu sync` or a
rename-migration command.

## Security boundary

The CLI rejects already tracked host files and blocks common secret-looking
paths by default. That is deliberately a guardrail, not data classification or
encryption. The trusted boundary is the user's private vault remote.
