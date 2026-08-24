# GitTrackUntracked (`gitu`)

> Track and sync selected untracked local project files into your own private Git
> vault without touching the original repo.

VS Code users can use the preview extension:
[GitTrackUntracked (Preview)](https://marketplace.visualstudio.com/items?itemName=MeetChaudhari.gittrackuntracked-vscode).

## The short version

Ever kept local notes, draft docs, TODOs, architecture decisions, scratch specs,
or research files beside a project because they were useful to you, but not meant
for a client/company repository?

GitTrackUntracked gives those explicitly selected, non-secret local files their
own private Git-backed vault, so your personal project context can follow you
across machines while the original repo stays clean.

## Why not just Git?

Git already has partial answers for parts of this workflow, but none of them are
quite the same thing:

| Existing option | Why it is not the same |
| --- | --- |
| `.git/info/exclude` | Hides files from local `git status`, but does not back them up or sync them to another machine. |
| Global gitignore | Useful for personal ignore rules, but not project-specific history or restore. |
| `git stash -u` | Good for temporary work, awkward as a long-term personal archive, and easy to lose in day-to-day repo work. |
| A local/private branch | Still mixes personal context with the project repository's branches and history, with a higher chance of accidental pushes or confusion. |
| An orphan branch | Possible, but heavy for this simple workflow and still lives inside the project repo. |
| A separate private repo | Works, but usually becomes manual copying and one-off structure per project. `gitu` turns that into one reusable private vault for many projects. |
| `assume-unchanged` / `skip-worktree` | Only applies to files already tracked by the host repo, not deliberately untracked personal files. |

The goal is not to replace Git. The goal is to make one small personal workflow
less fragile: explicitly selected local files stay out of the main repository,
but still get private history, sync, and restore.

## Why this exists

Have you ever made a useful local file for a project—perhaps
`docs/working-notes.md`, a draft client specification, a decision log, or a
scratch checklist—and deliberately kept it out of the main repository? It is
often the right choice: the file is useful to you, but it does not belong in a
client repository, a company repository, or production history.

Then you change laptops, set up a fresh checkout, work from another machine, or
switch to a branch where the same local context would help—and that file is
gone. You can copy it manually, but that turns into a private, error-prone
second workflow with no history.

GitTrackUntracked gives those *explicitly selected* local files a separate Git
history in one personal vault. Your normal repository remains exactly as it is;
your notes remain portable, versioned, and available only through the private
remote you choose.

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
| VS Code extension | Preview | Activity Bar dashboard and command-palette integration: [Marketplace link](https://marketplace.visualstudio.com/items?itemName=MeetChaudhari.gittrackuntracked-vscode). |
| npm CLI wrapper / JavaScript SDK | Experimental prerelease | Install the CLI with the `experimental` npm tag. |
| Conflict resolution / concurrent device edits | Not yet supported | Pull and resolve vault Git conflicts manually before retrying. |
| Encryption, access management, secret storage | Not supported | Use a secret manager. |
| Automatic Git hooks / background sync | Not supported | Deliberate manual sync keeps behavior visible. |

## Install

The simplest installation is the experimental npm package (Node.js 18+):

```sh
npm install --global gittrackuntracked@experimental
gitu --help
```

To build the CLI from source instead, install Go 1.26+ and run:

```sh
go install github.com/MeetChaudhari/GitTrackUntracked/cmd/gitu@latest
gitu --help
```

From a clone of this repository, use `go build -o bin/gitu ./cmd/gitu`.

## First-time setup: protect one local document

This is the complete workflow a new user can follow. Try it first with a
non-sensitive note in a test project.

### 1. Create one personal private vault

Create an empty **private** Git repository that you personally control. For
example, create `my-local-work-vault` on GitHub. Do **not** use a client,
company, or production repository for this vault.

Connect the machine to that vault once:

```sh
gitu vault init --remote git@github.com:you/my-local-work-vault.git
```

This clones the vault to your platform's local configuration directory
(`~/Library/Application Support/gitu/vault` on macOS; normally
`~/.config/gitu/vault` on Linux). Set `GITU_VAULT` before running `gitu` if you
prefer a different local location.

### 2. Register a deliberately local file in a project

Open a normal Git checkout and create the document you want to keep private to
your own vault—for example `docs/local-decisions.md`. It must be untracked by
the host repository.

```sh
cd path/to/your-project
gitu init
gitu add docs/local-decisions.md
gitu status
gitu sync -m "Save local decisions"
```

`gitu init` identifies a repository from its `origin` URL. If the project has
no `origin`, choose a stable ID and use exactly the same ID on another machine:

```sh
gitu init --project client-site
```

If you want normal Git to hide the local document from `git status`, add its
path to `.git/info/exclude`. That file is local to your checkout, unlike the
shared `.gitignore`, and `gitu` does not modify either file.

`gitu add` can also register a directory such as `personal-specs/`. It rejects
files already tracked by the host repository and blocks likely secret paths by
default. Do not use it for credentials, `.env` files, keys, or certificates.

### 3. Use the same vault with another project

You do not make a new private vault per project. In the second Git checkout,
run the same project setup and sync:

```sh
cd path/to/another-project
gitu init
gitu add docs/working-notes.md
gitu sync -m "Save notes for another project"
```

Both projects are stored independently inside the one vault.

### 4. Restore after changing machines

Install `gitu` on the new machine, connect it to the same vault, and then open
the same host project checkout:

```sh
gitu vault init --remote git@github.com:you/my-local-work-vault.git
cd path/to/your-project
gitu init
gitu restore
```

When another machine may have pushed newer changes, run `gitu vault pull`
before `restore` or `sync`. `restore` never overwrites an existing destination
unless you explicitly add `--force`.

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

- [`gittrackuntracked`](https://www.npmjs.com/package/gittrackuntracked) — npm-installed native CLI.
- [`gittrackuntracked-sdk`](https://www.npmjs.com/package/gittrackuntracked-sdk) — JavaScript adapter around the CLI.
- [`MeetChaudhari.gittrackuntracked-vscode`](https://marketplace.visualstudio.com/items?itemName=MeetChaudhari.gittrackuntracked-vscode) — VS Code Activity Bar dashboard and command-palette integration.

Install the npm CLI with `npm install --global gittrackuntracked@experimental`.

### VS Code dashboard

After installing the extension and opening a project folder, select the
**GitTrackUntracked** icon in the Activity Bar. The **Private Local Files**
view shows the vault/project state, registered paths, and actions to initialize,
add, sync, restore, and refresh. It calls the same local `gitu` CLI; configure
`gittrackuntracked.binaryPath` in VS Code Settings if `gitu` is not on `PATH`.

## Contributing, support, and security

Issues and pull requests are welcome while the project is experimental. Start with [CONTRIBUTING.md](CONTRIBUTING.md), use the included GitHub issue forms, and follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Do not report security issues in public issues; use [SECURITY.md](SECURITY.md) instead.

## Credits and license

MIT licensed. Initial credits are in [AUTHORS.md](AUTHORS.md); replace or expand them with the maintainers' preferred public names before the first public push.
