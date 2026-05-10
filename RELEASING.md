# Releasing

## Version scheme

`v<go-googlesql-Major.Minor>.<YYYYMMDD>`.

- `Major.Minor` follow the linked `go-googlesql` release. When
  `go-googlesql` is at `v0.1.x`, this project releases as
  `v0.1.YYYYMMDD`. Bump to `v0.2.YYYYMMDD` when upstream moves to
  `v0.2.x`.
- The patch component is the release date in `YYYYMMDD`.
- If a tag for that date already exists, advance the date by one until
  a free slot is found (carry-over). Real-clock dates may be exceeded;
  the patch is just a strictly-monotonic free counter that happens to
  read as a date.

## Steps

1. Confirm `go.mod` is on the intended `go-googlesql` minor
   line. Update with `go get` if needed and re-run `mise run ci`.
2. Pin `third_party/googlesql` to the same upstream commit
   `go-googlesql` is built from (see the upstream `README.md`'s
   "Tracks GoogleSQL revision …" line). Commit any submodule bump.
3. Decide the patch date:
   - Default: today's UTC date.
   - If a tag already exists for that date, increment by one day until
     free.
4. Confirm CI is green on `main` for the commit you intend to release
   (see [CI verification](#ci-verification) below).
5. `git tag -a v<MAJOR>.<MINOR>.<YYYYMMDD>` and `git push origin <tag>`.
6. Confirm the **Release** workflow (`.github/workflows/release.yml`)
   completed successfully for that tag. It cross-builds `execute_query`
   for linux (amd64, arm64), darwin (amd64, arm64), and windows (amd64),
   uploads archives plus `SHA256SUMS` to the GitHub release, and names
   assets so [mise](https://mise.jdx.dev/dev-tools/backends/github.html)
   can pick the correct archive per OS/arch without extra configuration.
   (Protobuf sources are generated offline and committed; the workflow does not run `protoc`.)

### Installing via mise (`github:` backend)

After a release has assets attached:

```sh
mise use -g github:apstndb/go-googlesql-executequery@v0.2.YYYYMMDD
# or
mise install github:apstndb/go-googlesql-executequery@latest
```

Pre-built archives follow `execute_query_0.2.YYYYMMDD_linux_amd64.tar.gz` (and
similar); mise matches `linux` / `darwin` / `windows` and `amd64` /
`arm64` in the filename.

### Back-filling assets for an older tag

If a tag predates the Release workflow or assets failed to upload, run
**Actions → Release → Run workflow**, set **tag** to the existing tag
(e.g. `v0.2.20260509`), and re-upload to that GitHub release.

## CI verification

Releases are validated by the GitHub Actions workflow `.github/workflows/ci.yml`
(job `ci`, matrix `ubuntu-latest` and `macos-latest`).

- Open the **Actions** tab for the repository, select the **CI** workflow, and
  confirm the run for the tagged commit (or the `main` commit you intend to
  release) completed successfully on both OS rows.
- If a run failed from infra flake (runner disconnect, transient registry
  outage), use **Re-run failed jobs** on that workflow run. If only one matrix
  leg failed, **Re-run failed jobs** retries just the failed OS.
- Required checks: `mise run ci` (lint, unit tests, build), `mise run
  test:corpus`, and `mise run test:integration` must pass on each matrix OS
  before you publish the release or treat the tag as good.
