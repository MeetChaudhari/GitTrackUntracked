# Branching, branch renames, and project renames

## Default: shared project files

New projects use `shared` storage. Every Git branch reads and writes the same
vault location:

```text
projects/<project-id>/files/<registered-path>
```

Choose this mode for notes and documents that should not change with source
branches.

## Branch-scoped storage

Run this inside the host repository to isolate future files by checked-out Git
branch:

```sh
gitu branch enable
gitu sync
```

The vault now stores a branch folder derived from the branch name plus a hash:

```text
projects/<project-id>/branches/<branch-key>/files/<registered-path>
```

The hash prevents collisions between branch names that sanitize to the same
folder name. A detached `HEAD` is deliberately rejected: there is no stable
branch name to use.

`gitu branch enable` does **not** automatically copy old shared history into
every branch. The current branch starts its branch-scoped history on the next
sync. Restore shared files before enabling if you want that content as the
starting point.

## Renaming a Git branch

Git cannot notify an external vault after a branch name has changed. Rename the
Git branch, then explicitly move its vault folder:

```sh
git branch -m feature/old-name feature/new-name
gitu branch rename --from feature/old-name --to feature/new-name
```

The command refuses to overwrite an existing destination branch folder. It
creates and pushes a normal commit in the personal vault. Do not run it from two
machines concurrently.

## Renaming or moving the host project

`gitu init` normally calculates an identity from the host `origin` URL. If that
URL changes, an existing checkout keeps its stored identity, but a fresh clone
would calculate a different one. Choose one of these safe paths:

1. **Keep the old identity:** on every fresh clone, run
   `gitu init --project <existing-id>`. This is the least disruptive option.
2. **Move the vault identity:** in an initialized checkout, run:

   ```sh
   gitu project rename --to new-project-id
   ```

   Then use `gitu init --project new-project-id` in other checkouts.

`project rename` moves all shared and branch-scoped data under `projects/` and
pushes a vault commit. It refuses an existing destination identity.

## Current limitations

- Branch enablement is per checkout. A fresh checkout learns the storage mode
  from `project.json` when it initializes against an existing vault.
- Path registration is currently shared across a project's branches; only file
  contents and history are isolated. Per-branch registration lists are planned.
- Branch rename and project rename are manual by design. Automatic hooks would
  be less transparent and harder to recover when a rename is interrupted.
- Concurrent migration or sync operations require normal Git coordination.
