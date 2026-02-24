# cut-the-bs

`cut-the-bs` is a desktop app for people who tailor resumes and cover letters often.

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

## Release and Homebrew maintenance

See `HOMEBREW.md` for the release checklist, cask details, and workflow notes.
