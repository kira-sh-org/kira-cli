# Contributing to Kira

Thanks for your interest in contributing! This guide covers local development, building, testing, formatting, and installation details for contributors.

## Development Setup

```bash
make dev-setup   # download modules, tidy go.mod
```

## Building

```bash
make build       # produces ./kira
```

Print version info embedded at build time:

```bash
./kira version
```

## Testing

```bash
make test
make test-coverage
```

End-to-end tests:

```bash
bash kira_e2e_tests.sh
bash kira_e2e_tests.sh --keep   # preserve last test directory
```

After a `--keep` run, the latest directory can be inspected with:

```bash
latest=$(ls -dt e2e-test/test-kira-* | head -1)
ls -la "$latest"
```

## Formatting, Vetting, and Linting

All formatting and linting is handled by `golangci-lint` using the configuration in `.golangci.yml`. This provides a single source of truth for code style.

Install dev tools (first time):

```bash
make dev-setup   # installs golangci-lint and goreleaser
```

Format and organize imports (writes changes):

```bash
make fmt   # runs: golangci-lint run --fix
```

Verify formatting, lint, and tests:

```bash
make lint   # runs: golangci-lint run (includes formatting, vet, and all linters)
make check  # runs: lint + test
```

## Testing Releases Locally

Before creating a release, you can test the release process locally using GoReleaser's snapshot mode:

```bash
make release-snapshot   # Builds binaries for all platforms without creating a GitHub release
```

This creates artifacts in the `dist/` directory that you can inspect. Note: GoReleaser must be installed (included in `make dev-setup`).

## Installing the Built Binary

The `make install` target follows standard Unix conventions and honors `PREFIX` and `DESTDIR`.

- PREFIX: Where the software should live on the target system (default `/usr/local`). Binaries are installed to `$(PREFIX)/bin`.
- DESTDIR: A temporary staging root prepended during install. Useful for packaging or sandboxed installs. It does not change the intended final location.

Common scenarios:

- Install for your user (no sudo):
  ```bash
  make install PREFIX="$HOME/.local"
  # Result: ~/.local/bin/kira
  ```

- System-wide install (default prefix):
  ```bash
  sudo make install
  # Result: /usr/local/bin/kira
  ```

- Staged install for packaging:
  ```bash
  make install PREFIX=/usr/local DESTDIR=/tmp/pkgroot
  # Files land in: /tmp/pkgroot/usr/local/bin/kira
  # Intended final path in the package: /usr/local/bin/kira
  ```

## Project Structure (Quick Reference)

```
cmd/kira/            # CLI entrypoint
internal/            # Commands, config, templates, validation
templates/           # Work item templates
Makefile             # Dev/build tooling
```

## Running

After building, test the binary:

```bash
./kira --help
```

## Contribution Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes (with tests)
4. Run `make check`
5. Open a pull request


