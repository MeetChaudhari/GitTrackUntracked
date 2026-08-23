# Contributing

Thank you for helping improve an experimental project.

## Before opening an issue

Search existing issues first. Never include vault URLs with credentials, private documents, `.env` contents, SSH keys, tokens, client names, or a copy of a private vault in an issue.

## Bug reports

Use the bug-report form and include the CLI version, operating system, Git version, sanitized command/output, expected behavior, and a minimal reproducible example. State whether the project used shared or branch-scoped storage.

## Pull requests

1. Keep each change focused.
2. Add or update tests for behavior changes.
3. Run `go test ./...`, `go vet ./...`, and `gofmt -w` on modified Go files.
4. Update the feature matrix and relevant documentation when behavior changes.
5. Do not change the experimental warning or claim a capability is supported until it is implemented and tested.

By contributing, you agree that your contributions may be released under the MIT License.
