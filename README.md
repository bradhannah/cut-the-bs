# cut-the-bs

Desktop app for tailoring resumes and cover letters (Wails + Go + Svelte).

## Install on macOS with Homebrew (this repo tap)

```bash
brew tap bradhannah/cut-the-bs https://github.com/bradhannah/cut-the-bs
brew install --cask cut-the-bs
```

Note: this expects a published GitHub Release asset named
`cut-the-bs-macos-universal.tar.gz`.

This app is currently unsigned. The Homebrew cask includes a post-install
quarantine removal step (`xattr`) to make first launch easier.

If macOS still blocks launch:

```bash
xattr -dr com.apple.quarantine "/Applications/cut-the-bs.app"
```

Upgrade to the latest release:

```bash
brew upgrade --cask --greedy cut-the-bs
```

Uninstall and remove app data:

```bash
brew uninstall --cask cut-the-bs
brew zap cut-the-bs
```

## Developer setup on macOS

```bash
brew bundle --file Brewfile
make setup
make dev
```

## Release + Homebrew maintenance

See `HOMEBREW.md` for the full release checklist and automation details.
