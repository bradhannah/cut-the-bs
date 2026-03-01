# cut-the-bs

<p align="center">
  <img src="assets/cut-the-bs.png" alt="Cut the BS app icon" width="180" />
</p>

<p align="center">
  <a href="https://github.com/bradhannah/cut-the-bs/releases"><img src="https://img.shields.io/github/v/release/bradhannah/cut-the-bs" alt="Release" /></a>
  <a href="https://github.com/bradhannah/cut-the-bs/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/bradhannah/cut-the-bs/ci.yml?label=build" alt="Build" /></a>
  <a href="https://github.com/bradhannah/cut-the-bs/releases"><img src="https://img.shields.io/github/downloads/bradhannah/cut-the-bs/total" alt="Downloads" /></a>
  <a href="https://github.com/bradhannah/cut-the-bs/stargazers"><img src="https://img.shields.io/github/stars/bradhannah/cut-the-bs" alt="Stars" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="MIT License" /></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Wails-2.11.0-2F67F6" alt="Wails" />
  <img src="https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/Svelte-3-orange?logo=svelte" alt="Svelte" />
  <img src="https://img.shields.io/badge/TypeScript-strict-blue?logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Bun-1.x-black?logo=bun" alt="Bun" />
  <img src="https://img.shields.io/badge/platform-macOS%20Universal-lightgrey" alt="Platform" />
  <img src="https://img.shields.io/badge/Homebrew-available-brown?logo=homebrew" alt="Homebrew" />
</p>

"Cut the Bullsh*t" (`cut-the-bs`) is a desktop app for people who tailor resumes and cover letters often.

It is called that because it helps you cut the unimportant bullshit from your resume and cover letter so hiring teams can see the real you.

It is built for practical job-search work: keep your experience data in one place, target each application cleanly, and generate role-specific documents without copy-paste chaos.

<p>
  <img src="docs/assets/readme/workflow-overview.svg" alt="Five-step workflow overview for cut-the-bs" width="980" />
</p>

## Who this is for

- Engineers, architects, consultants, and technical leads applying across multiple related roles.
- People with deep work history who need selective, role-specific storytelling.
- Anyone who wants local-first data ownership (SQLite on your machine, no web account required).

## Why people use it

- Job applications become the center of the workflow, so each resume/cover letter pair stays tied to a specific role.
- Lenses let you reuse your core experience but change emphasis quickly.
- Templates are customizable enough to support both conservative and modern formats.
- Versioned exports make it easy to iterate instead of rewriting from scratch every time.

## What it does

- **Experience source of truth**: profile, work history, bullets, skills, summaries, certs, and more.
- **Application tracking**: status timeline, fit indicator, notes, posting URL, and linked document versions.
- **Resume + cover letter generation**: create new versions or overwrite latest per application.
- **Template builder**: built-in templates (view-only) plus user templates you can duplicate and edit.
- **Data controls**: configurable data directory, full export/import, and rolling backups.

<p>
  <img src="docs/assets/readme/screen-map.svg" alt="Map of the main cut-the-bs app sections" width="980" />
</p>

## Screenshots

Here are real screens from the app so people can see what daily use actually looks like.

### Applications-first workflow

<p>
  <img src="docs/assets/applications.png" alt="Applications view with statuses and document actions" width="980" />
</p>

### Build a targeted resume

<p>
  <img src="docs/assets/build-a-resume.png" alt="Resume generation setup with lens and template selection" width="980" />
</p>

### Lenses for role-specific emphasis

<p>
  <img src="docs/assets/lenses.png" alt="Lens management and selection controls" width="980" />
</p>

### Skills depth and categorization

<p>
  <img src="docs/assets/skills.png" alt="Skills page with categories and competence levels" width="980" />
</p>

### Example generated resume output

<p>
  <img src="docs/assets/resume-snap.png" alt="Generated resume PDF preview" width="900" />
</p>

## Install on macOS (Homebrew)

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

Note: the app is currently unsigned. If Gatekeeper blocks launch:

```bash
xattr -dr com.apple.quarantine "/Applications/cut-the-bs.app"
```

## Screenshot-safe sample data

You can point the app to a separate data directory for demos/screenshots.

- Open `Settings` -> `Data Management`.
- Set `Data Directory` to your sample folder.
- Save and restart.

This keeps personal data and demo data fully separate.

## Developer setup (macOS)

```bash
brew bundle --file Brewfile
make setup
make dev
```

## ATS compatibility checks

You can validate both generated resumes and cover letters with automated tests and a local PDF checker.

Run ATS integration tests:

```bash
go test ./tests/integration -run ATS -count=1
```

Or use the Make target:

```bash
make test-ats
```

Run the local checker against a real exported PDF:

```bash
go run ./cmd/atscheck --pdf "/path/to/generated.pdf"
```

Add expected content and reading-order checks:

```bash
go run ./cmd/atscheck \
  --pdf "/path/to/generated.pdf" \
  --require "Jane Smith" \
  --require "Acme Robotics" \
  --order "Jane Smith>Experience>Acme Robotics"
```

Shortcut Make targets:

```bash
make atscheck PDF="/path/to/generated.pdf"
make atscheck-resume PDF="/path/to/resume.pdf"
make atscheck-cover-letter PDF="/path/to/cover-letter.pdf"
```

You can append extra checks to Make targets with `ATS_ARGS`, for example:

```bash
make atscheck-resume PDF="/path/to/resume.pdf" ATS_ARGS='--require "Acme Robotics" --order "Summary>Experience"'
```

## Release and Homebrew maintenance

See `HOMEBREW.md` for the release checklist, cask details, and workflow notes.
