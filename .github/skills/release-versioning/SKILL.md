---
name: release-versioning
description: End-to-end runbook for cutting a new cut-the-bs release and updating the Homebrew cask safely. Use this when asked to publish a new version, bump cask version and checksum, or troubleshoot Homebrew SHA-256 mismatch errors.
license: MIT
---

You are responsible for release and versioning tasks in this repository. Follow this runbook exactly to keep GitHub releases, workflow outputs, and Homebrew cask metadata in sync.

## Scope

This skill covers:

- Creating a new immutable release tag.
- Waiting for the tag-triggered GitHub Actions release workflow.
- Reading the authoritative checksum from the published release asset.
- Updating `Casks/cut-the-bs.rb` with the new `version` and `sha256`.
- Verifying `brew upgrade` works end-to-end.
- Handling checksum mismatch incidents safely.

## Repository anchors

- Release workflow: `.github/workflows/release-macos.yml`
- Release trigger: push tag matching `v*`
- Build/package command used by workflow: `make release-macos`
- Release asset: `cut-the-bs-macos-universal.tar.gz`
- Homebrew cask: `Casks/cut-the-bs.rb`
- Tap install target: `bradhannah/cut-the-bs/cut-the-bs`

## Required guardrails

1. Treat tags and release assets as immutable.
2. Never overwrite assets for an already-published version.
3. If any artifact differs after publication, release a new patch version.
4. Use the checksum from GitHub release asset metadata as the source of truth.
5. Keep cask `version` and release tag aligned (`v#{version}`).
6. Do not use force-push or retagging for releases.

## Standard release workflow

Set variables:

```bash
REPO="bradhannah/cut-the-bs"
VERSION="0.1.X"
TAG="v${VERSION}"
ASSET="cut-the-bs-macos-universal.tar.gz"
```

### 1) Preflight checks

```bash
git checkout main
git pull --ff-only origin main
git status --short --branch
git tag -l "${TAG}"
git ls-remote --tags origin "${TAG}"
```

Rules:

- Continue only if working tree is clean.
- Continue only if the tag does not exist locally or remotely.

### 2) Create and push tag

```bash
git tag "${TAG}"
git push origin "${TAG}"
```

### 3) Wait for release workflow to complete

```bash
gh run list --repo "${REPO}" --workflow release-macos.yml --limit 5
# Identify the run id for ${TAG}, then:
gh run watch <run-id> --repo "${REPO}" --exit-status
```

The release is valid only if the run concludes with `success`.

### 4) Read authoritative checksum from release asset

```bash
gh release view "${TAG}" --repo "${REPO}" --json assets
```

Find the asset named `cut-the-bs-macos-universal.tar.gz` and copy its `digest` value (format: `sha256:<hex>`).

Extract raw checksum (without prefix):

```bash
DIGEST="$(gh release view "${TAG}" --repo "${REPO}" --json assets --jq ".assets[] | select(.name==\"${ASSET}\") | .digest")"
SHA256="${DIGEST#sha256:}"
echo "${SHA256}"
```

Optional integrity cross-check:

```bash
curl -L "https://github.com/${REPO}/releases/download/${TAG}/${ASSET}" | shasum -a 256
```

The computed value must match `SHA256` exactly.

### 5) Update Homebrew cask

Edit `Casks/cut-the-bs.rb`:

- Set `version` to `${VERSION}`.
- Set `sha256` to `${SHA256}`.
- Keep URL pattern unchanged.

### 6) Commit and push cask bump

Use repository commit style:

```bash
git add Casks/cut-the-bs.rb
git commit -m "chore: pin Homebrew cask to ${TAG} release artifact"
git push origin main
```

### 7) Validate Homebrew upgrade path

```bash
brew update
brew upgrade --cask bradhannah/cut-the-bs/cut-the-bs
brew info --cask bradhannah/cut-the-bs/cut-the-bs
brew outdated --cask cut-the-bs
brew upgrade --cask bradhannah/cut-the-bs/cut-the-bs
```

Expected results:

- First upgrade succeeds without checksum errors.
- `brew outdated --cask cut-the-bs` returns no entry.
- Second upgrade reports latest version already installed.

## Incident playbook: SHA-256 mismatch

When Homebrew reports:

- `Expected` != `Actual`

Run diagnostics:

```bash
shasum -a 256 "/path/from/homebrew/error/cached-file.tar.gz"
curl -L "https://github.com/${REPO}/releases/download/${TAG}/${ASSET}" | shasum -a 256
gh release view "${TAG}" --repo "${REPO}" --json assets
```

Interpretation:

- If `curl` hash equals Homebrew `Actual`, cask checksum is stale.
- If release digest changed for an already published version, do not patch old version in place. Cut a new patch release and update cask to that new version.

User recovery command after fix:

```bash
rm "/Users/<user>/Library/Caches/Homebrew/downloads/<file-from-error>"
brew update
brew upgrade --cask bradhannah/cut-the-bs/cut-the-bs
```

## Lessons learned to enforce

- A mutable release asset under the same version breaks Homebrew installs.
- Workflow reruns or manual asset replacement can change checksums.
- Cask checksum must always be tied to the final published asset digest.
- Version bumps are cheaper and safer than trying to preserve mutable artifacts.

## Completion checklist

Before declaring release done, confirm all are true:

- Tag exists remotely and points to intended commit.
- Release workflow is green for that tag.
- Release asset checksum is recorded and verified.
- Cask `version` and `sha256` match the release asset.
- Cask change is committed and pushed to `main`.
- Local Homebrew upgrade test succeeds end-to-end.
