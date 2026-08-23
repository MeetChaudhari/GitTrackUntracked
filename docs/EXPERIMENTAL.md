# Experimental status and risk boundary

`gitu` is an early public experiment. It should be evaluated with a disposable
vault and copies of non-critical documents before anyone trusts it with routine
work.

## What the experimental label means

- Command names, the vault layout, and configuration format can change before
  1.0.
- We may require a migration between preview releases.
- Error recovery, concurrent editing, Windows behavior, unusual filesystems,
  and large-file performance have not received the breadth of testing required
  for a stable release.
- Git writes are real. `sync`, branch migration, and project migration commit
  to and push from the personal vault.

## Non-goals

This tool does not encrypt files, manage user access, scan for every secret,
replace backups, or replace a secret manager. A private remote is an access
control decision made by its owner—not a guarantee provided by `gitu`.

## Required user practices

1. Use a separate private vault remote, not an employer or client remote.
2. Do not store credentials, `.env` files, private keys, certificates, or
   regulated data. The path guard is helpful but not exhaustive.
3. Keep an independent backup until the project is stable.
4. Run `gitu vault pull` before working on a second machine.
5. Inspect and resolve any vault Git conflict directly; do not delete the vault
   directory to “fix” one.

## Stable-release exit criteria

We will not remove this warning until versioned migrations, recovery guidance,
cross-platform CI, concurrency behavior, and independent user testing are in
place. The maintainers track release readiness privately.
