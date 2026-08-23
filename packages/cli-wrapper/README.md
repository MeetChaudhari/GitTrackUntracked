# `gittrackuntracked` (preview)

This package downloads the matching GitTrackUntracked binary from the matching
GitHub prerelease and verifies it against the release checksum manifest.

For local development, point it at a built native executable:

```sh
GITU_BINARY="$PWD/bin/gitu" node packages/cli-wrapper/bin/gitu.cjs --help
```

Supported targets are macOS and Linux on x64/arm64, plus Windows on x64/arm64.
