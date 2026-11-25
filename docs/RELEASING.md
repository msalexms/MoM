# Releasing mom

`mom` follows [Semantic Versioning](https://semver.org/). Releases are built and
published by GoReleaser from a Git tag — nothing is uploaded by hand.

While the project is pre-1.0, a **minor** bump (`0.x.0`) carries new features and a
**patch** bump (`0.x.y`) is reserved for fixes and packaging changes.

## Before you tag

Open a `release/vX.Y.Z` branch and complete this checklist:

1. **CI is green** on `main` — the `CI` workflow must pass before branching.
2. **Changelog** — in [`CHANGELOG.md`](../CHANGELOG.md):
   - rename `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD`;
   - add the `[X.Y.Z]` compare link at the bottom;
   - open a fresh, empty `## [Unreleased]` section.
3. **Readme** — in both [`README.md`](../README.md) and
   [`README.es.md`](../README.es.md), update the version line in the
   **What's new / Novedades** section and refresh the bullets.
4. **Counts and claims** — confirm the module, theme, template, and style numbers
   still match the code (`internal/module`, `internal/theme`, `embed/templates`).
5. **Local check**:

   ```bash
   make fmt
   make lint
   make test
   make build
   ./bin/mom --version
   ```

   The build targets `linux/amd64`; on Windows or macOS rely on CI for the binary
   smoke test.

6. **Commit and merge** — commit as `chore(release): vX.Y.Z`, open the PR, and
   merge once CI passes.

## Tag and publish

```bash
git switch main
git pull --ff-only
git tag -a vX.Y.Z -m "mom vX.Y.Z"
git push origin vX.Y.Z
```

Pushing the tag triggers [`release.yml`](../.github/workflows/release.yml), which
runs GoReleaser and publishes:

| Artifact | Name |
| --- | --- |
| Linux archives | `mom_linux_amd64.tar.gz`, `mom_linux_arm64.tar.gz`, `mom_linux_armv7.tar.gz` |
| Debian / Ubuntu | `mom_X.Y.Z_amd64.deb`, `mom_X.Y.Z_arm64.deb` |
| Fedora / RHEL | `mom_X.Y.Z_amd64.rpm`, `mom_X.Y.Z_arm64.rpm` |
| Checksums | `checksums.txt` |
| Release notes | Generated from commits, grouped by type |

Archive names are intentionally **unversioned** so the stable links keep working:

```bash
curl -sL https://github.com/msalexms/MoM/releases/latest/download/mom_linux_amd64.tar.gz | tar xz
sudo install -m755 mom /usr/local/bin/mom
mom --version
```

## After the release

- Check the release page: notes grouped under *Features*, *Fixes*, and
  *Performance*, with the archives, packages, and checksums attached.
- Run the stable one-liner above on a Linux machine and confirm `mom --version`
  prints the tag, commit, and build date.
- Announce the release (and update the Spanish release notes if needed).
- Leave `## [Unreleased]` empty and ready for the next cycle.

## Commit convention

Release notes are generated from commit subjects, so keep them conventional:

```text
feat(module): add a new module
fix(tui): keep the cursor on screen
perf(generator): render sections in parallel
refactor(config): simplify profile loading
docs: explain the login hook
chore: bump dependencies
ci: cache Go modules
```

GoReleaser groups `feat`, `fix`, and `perf` and hides `docs`, `chore`, `ci`,
`style`, and `test`.

## If a release goes wrong

- **Workflow failed before publishing**: delete the tag locally and remotely
  (`git tag -d vX.Y.Z`, `git push origin :vX.Y.Z`), fix the problem, and tag again.
- **Release already published**: ship a patch release instead. Never move or
  force-push a tag that users may already have downloaded.
