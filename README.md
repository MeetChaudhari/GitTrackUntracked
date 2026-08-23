# GitTrackUntracked (`gitu`)

> **EXPERIMENTAL — DO NOT USE IN PRODUCTION.**
>
> `gitu` is pre-1.0 software being prepared for an initial public release. It can create and push real Git history. Keep independent backups, use a test vault first, and do not rely on it for critical or regulated data.

`gitu` gives deliberately untracked project files a private, portable Git home. It is for local notes, specifications, work plans, scratch documents, and other project-adjacent files that should follow you between machines but must never be added to a client, company, or production repository.

## One private vault, every project

Create **one private Git repository** that you control—on GitHub, GitLab, a self-hosted server, or elsewhere. `gitu` clones it locally once and namespaces each host project within it:

```text
my-local-work-vault/
└── projects/
    ├── repo-a1b2c3d4e5f6/
    │   ├── project.json
    │   ├── manifest.json
    │   └── files/docs/local-decisions.md
    └── repo-0f1e2d3c4b5a/
        └── files/notes.md
```

The company/client repository is not changed. Its only `gitu` state is held inside that repository's `.git` directory, so `git status`, `.gitignore`, host commits, remotes, hooks, and collaborators are unaffected.

## Feature matrix

| Capability | Status in `0.1.0-experimental` | Notes |
| --- | --- | --- |
| One private vault for many projects | Supported | Each project is namespaced under `projects/`. |
| Explicit opt-in paths | Supported | `gitu` never scans or backs up all untracked files. |
| Files and directories | Supported | Copies regular files, directories, permissions, and symbolic links. |
| Reject host-tracked files | Supported | Includes tracked files found inside an added directory. |
| Secret-like path guard | Supported | `.env`, keys, certificates, credential-like names, and `.aws` require an override. |
| Sync/commit/push | Supported | Uses ordinary Git commits in the vault. |
| Restore on a new machine | Supported | `init` imports the saved manifest, then `restore` copies files back. |
| Branch-scoped vault files | Supported, experimental | Enable per checkout with `gitu branch enable`. |
| Branch rename migration | Supported, manual | Run `gitu branch rename --from OLD --to NEW` after a Git rename. |
| Project/repository rename migration | Supported, manual | Run `gitu project rename --to NEW-ID`; see the rename guide. |
| VS Code extension | Preview scaffold | Commands are present; marketplace release is not ready. |
| npm CLI wrapper / JavaScript SDK | Preview scaffold | Not published until platform binaries and release checks exist. |
| Conflict resolution / concurrent device edits | Not yet supported | Pull and resolve vault Git conflicts manually before retrying. |
| Encryption, access management, secret storage | Not supported | Use a secret manager. |
| Automatic Git hooks / background sync | Not supported | Deliberate manual sync keeps behavior visible. |

## Install the experimental CLI

Requires Go 1.26+ to build from source:

```sh
go install github.com/MeetChaudhari/GitTrackUntracked/cmd/gitu@latest
```

From this checkout:

```sh
go build -o bin/gitu ./cmd/gitu
```

## Quick start

First, create an empty **private** repository yourself. Then, once per machine:

```sh
gitu vault init --remote git@github.com:you/my-local-work-vault.git
```

Inside a normal Git working tree:

```sh
gitu init
gitu add docs/local-decisions.md personal-specs/
gitu status
gitu sync -m "Capture current decisions"
```

`gitu init` derives a stable identity from the host repository's `origin` URL. For a project without an origin, choose an identity that you will use again on other machines:

```sh
gitu init --project client-site
```

## Restore on another machine

```sh
gitu vault init --remote git@github.com:you/my-local-work-vault.git
cd client-project
gitu init
gitu restore
```

`restore` does not overwrite an existing destination unless you supply `--force`. Run `gitu vault pull` first if another machine could have pushed newer changes.

## Branches and renamed projects

By default, a project's files are shared by all branches. When the untracked files should diverge with source branches, opt into branch-scoped storage:

```sh
gitu branch enable
gitu sync
```

After Git renames a branch, move its vault history explicitly:

```sh
git branch -m feature/docs feature/notes
gitu branch rename --from feature/docs --to feature/notes
```

The detailed behavior, migration steps, and limits are in [docs/BRANCHING_AND_RENAMES.md](docs/BRANCHING_AND_RENAMES.md).

## Command reference

| Command | Purpose |
| --- | --- |
| `gitu vault init --remote URL` | Clone or reconnect the one personal vault. |
| `gitu vault pull` | Fast-forward the local vault before restoring or syncing. |
| `gitu init [--project ID]` | Link a checkout to a vault project identity. |
| `gitu add [--allow-sensitive] PATH...` | Register an untracked file or directory. |
| `gitu list` | Show paths registered in this checkout. |
| `gitu status` | Compare registered paths with the local vault. |
| `gitu sync [-m MESSAGE]` | Copy, commit, and push this project's selected paths. |
| `gitu restore [--force] [PATH...]` | Restore selected paths from the vault. |
| `gitu branch enable` | Isolate future vault files per Git branch. |
| `gitu branch status` | Show whether branch-scoped storage is active. |
| `gitu branch rename --from OLD --to NEW` | Move a branch's vault directory after a Git rename. |
| `gitu project rename --to ID` | Move all vault history to a new stable project identity. |

Set `GITU_VAULT` to choose a non-default local location for the cloned vault.

## Safety and known limitations

- `gitu` is not a secrets manager and does not encrypt content.
- `--allow-sensitive` is an intentional override, not a security guarantee.
- A private Git remote can still be copied, compromised, or misconfigured.
- A failed push leaves a local vault commit; inspect it with Git and retry after fixing the remote issue.
- There is no automatic merge strategy for concurrent edits from multiple machines. Pull, resolve the normal Git conflict in the vault, then rerun.
- Moving a vault directory via `branch rename` or `project rename` happens locally first and then pushes a normal Git commit. Do not interrupt it.

Read [docs/EXPERIMENTAL.md](docs/EXPERIMENTAL.md) before adopting the tool.

## Ecosystem packages

This repository includes experimental distribution packages. They remain
prerelease software and are not suitable for production use.

- [`@meetchaudhari/gittrackuntracked`](https://www.npmjs.com/package/@meetchaudhari/gittrackuntracked) — npm-installed native CLI.
- [`@meetchaudhari/gittrackuntracked-sdk`](https://www.npmjs.com/package/@meetchaudhari/gittrackuntracked-sdk) — JavaScript adapter around the CLI.
- [`MeetChaudhari.gittrackuntracked-vscode`](https://marketplace.visualstudio.com/items?itemName=MeetChaudhari.gittrackuntracked-vscode) — VS Code command-palette integration.

Install the npm CLI with `npm install --global @meetchaudhari/gittrackuntracked@experimental`.

## Contributing, support, and security

Issues and pull requests are welcome while the project is experimental. Start with [CONTRIBUTING.md](CONTRIBUTING.md), use the included GitHub issue forms, and follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Do not report security issues in public issues; use [SECURITY.md](SECURITY.md) instead.

## Credits and license

MIT licensed. Initial credits are in [AUTHORS.md](AUTHORS.md); replace or expand them with the maintainers' preferred public names before the first public push.
