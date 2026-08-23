# `@gittrackuntracked/cli` (preview)

This package is the planned npm distribution wrapper for the Go `gitu` binary.
It resolves a platform-specific optional package once those signed binaries are
released. Until then it is source-only and intentionally not publishable as a
working end-user installer.

For local development, point it at a built native executable:

```sh
GITU_BINARY="$PWD/bin/gitu" node packages/cli-wrapper/bin/gitu.cjs --help
```

Do not claim `npx @gittrackuntracked/cli` support before the release checklist is
complete.
