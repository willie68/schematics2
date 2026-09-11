---
name: release
description: >
  Release workflow for the schematics2 project only: run tests, bump versions,
  commit/push to GitHub, then Docker build via buildDocker.cmd. Use when the user
  asks to release, ship, or deploy schematics2 (not other repos).
disable-model-invocation: true
---

# Release (schematics2 only)

Repo root: schematics2 (this project). Do **not** apply to other workspaces.

Stop the whole release on any failed step. Confirm with the user before version bump type, commit, and Docker deploy side effects.

## Progress

```
- [ ] 1. Tests
- [ ] 2. Version bump
- [ ] 3. GitHub check-in
- [ ] 4. Docker build (buildDocker.cmd)
```

---

## Step 1 — Alle Tests laufen lassen

Working directory: `backend/`

```bash
go test ./...
```

- Exit non-zero → stop. Report failures. No version bump, no commit, no build.
- Frontend has no `npm test` today. If a frontend test script appears later, run it from `frontend/` as well before continuing.
- Optional sanity: `go build ./...` in `backend/` if tests pass but build risk looks high.

---

## Step 2 — Versionsnummer erhöhen (nur wenn Tests OK)

Versions are **independent** and may diverge:

| Component | Source of truth |
|---|---|
| Backend | `backend/internal/version/version.go` → `Version = "x.y.z"` |
| Frontend | `frontend/package.json` → `"version": "x.y.z"` |

Also keep in sync when bumping:

- `frontend/package-lock.json` top-level `version` (and the package entry `version`) to match `package.json`
- `HISTORY.md` — prepend a new section at the top (after `# History`)

### Decide what to bump

From the uncommitted / release scope, ask the user if unclear:

1. **Which side(s)?** `backend` | `frontend` | `both`
2. **Semver level?** default **patch** (`0.3.9` → `0.3.10`) unless user says minor/major

Read current versions from both files before editing. Show old → new and wait for confirmation.

### HISTORY.md format

```markdown
## x.y.z - YYYY-MM-DD (Backend | Frontend | Backend & Frontend)

- Short bullets of what ships in this release
```

Use today's date. Mark which side(s) changed in the heading.

Ensure `HISTORY.md` has **no merge conflict markers** before committing.

---

## Step 3 — Check-in GitHub

Only after steps 1–2 succeed and the user confirms the commit message.

1. `git status` / `git diff` — stage release-related files (version sources, lockfile, HISTORY, and any other intended release changes).
2. Commit (imperative title), e.g.:

```
release: bump backend/frontend to 0.3.10
```

Or side-specific:

```
release: bump frontend to 0.3.10
```

3. Push to `origin` on the current branch:

```bash
git push -u origin HEAD
```

Remote is typically `https://github.com/willie68/schematic.git`. Do **not** force-push.

If nothing to commit after the bump, still ensure remote is up to date before step 4.

---

## Step 4 — Build laufen lassen (`buildDocker.cmd`)

Run **after** the GitHub check-in so the image’s `VCS_REF` matches the pushed commit.

Working directory: `backend/scripts/`

```bat
buildDocker.cmd
```

Default `BASE_PATH` is `/schematics2` (reverse-proxy). Only pass another path if the user asks (e.g. `buildDocker.cmd /client`).

This script:

1. Builds Docker image `mcs/schematics2:latest`
2. Tags/pushes to `192.168.178.14:5000/mcs/schematics2:latest`
3. SSHs to the server and runs `/opt/schematics2/update.sh`

Needs Docker + network + SSH access. If it fails mid-way, report which step failed; do not re-bump versions.

---

## Rules

- Scope: **schematics2 only**
- Never skip tests
- Never invent version numbers — read files, then bump
- Never commit without explicit user OK on the message
- Never run `buildDocker.cmd` before push when a new commit was created
- Ask before minor/major or one-sided bumps when the change set is ambiguous
