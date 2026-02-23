# Homebrew Distribution Plan (Single Repo)

This project distributes the macOS app from the same repository (no separate
`homebrew-...` tap repo).

## What this setup provides

- Developer bootstrap via `Brewfile` (`brew bundle`).
- End-user install via Homebrew cask from this repository (`Casks/cut-the-bs.rb`).
- Release artifact packaging in `make release-macos`.
- CI release automation in `.github/workflows/release-macos.yml`.
- Unsigned-app convenience via cask `postflight` quarantine removal (`xattr`).

## User install commands

```bash
brew tap bradhannah/cut-the-bs https://github.com/bradhannah/cut-the-bs
brew install --cask cut-the-bs
```

Upgrade:

```bash
brew upgrade --cask --greedy cut-the-bs
```

Uninstall:

```bash
brew uninstall --cask cut-the-bs
brew zap cut-the-bs
```

## Release checklist (copy/paste)

Run from `main` branch with a clean tree.

1) Validate baseline quality via Makefile

```bash
make test
make frontend-check
```

2) Build + package universal macOS app

```bash
make release-macos
```

This generates:

- `dist/cut-the-bs-macos-universal.tar.gz`
- `dist/cut-the-bs-macos-universal.tar.gz.sha256`

3) Commit release prep (if any)

```bash
git add -A
git commit -m "chore(release): prepare vX.Y.Z"
```

4) Tag and push

```bash
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

5) Verify GitHub release artifact

- Workflow `.github/workflows/release-macos.yml` runs on `v*` tags.
- It uploads `cut-the-bs-macos-universal.tar.gz` and checksum file to the release.

6) Smoke test install on another Mac

```bash
brew tap bradhannah/cut-the-bs https://github.com/bradhannah/cut-the-bs
brew install --cask cut-the-bs
open /Applications/cut-the-bs.app
```

If Gatekeeper still blocks launch:

```bash
xattr -dr com.apple.quarantine "/Applications/cut-the-bs.app"
```

## Notes on unsigned app behavior

- The cask includes a `postflight` step to remove quarantine recursively:
  - `/usr/bin/xattr -dr com.apple.quarantine "/Applications/cut-the-bs.app"`
- This improves first-run UX for internal/test distribution.
- For production-quality public distribution, add signing + notarization later.
