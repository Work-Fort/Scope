# Go Frontend SDK — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restore the `pkg/frontend` Go package as `go/frontend/`, a standalone Go module that other WorkFort services import to serve federated UI modules through Scope.

**Architecture:** The package lives at `go/frontend/` with its own `go.mod` (module path `github.com/Work-Fort/Scope/go/frontend`). It has zero external dependencies — only stdlib. A GitHub Actions workflow tags and releases the module on push to master, following the same pattern as Passport's `release-sdk-go.yml`. The docs are updated to reference the new import path.

**Tech Stack:** Go 1.25 (via mise), GitHub Actions, `Work-Fort/github-tag-action`

**References:**
- Git history: `437de6a` has the last version of `pkg/frontend/frontend.go` and `pkg/frontend/frontend_test.go`
- Passport pattern: `~/Work-Fort/Passport/lead/.github/workflows/release-sdk-go.yml`
- Passport Go SDK: `~/Work-Fort/Passport/lead/go/service-auth/`
- Passport mise tasks: `~/Work-Fort/Passport/lead/.mise/tasks/sdk/go/{build,test}`
- Scope mise tasks: `.mise/tasks/` (file-based, nested dirs like `web/{build,dev}`)
- Docs to update: `docs/frontend/service-contract.md`, `docs/frontend/architecture.md`, `docs/frontend/getting-started/solidjs.md`, `docs/frontend-docs-design.md`

---

### Task 1: Add Go to mise tooling and create go.mod

**Files:**
- Modify: `mise.toml`
- Create: `go/frontend/go.mod`

**Step 1: Add Go to mise.toml**

Add `go = "1.25"` to the `[tools]` section in `mise.toml`.

**Step 2: Install Go via mise**

Run: `mise install`
Expected: Go 1.25 installed and available

**Step 3: Create the module file**

```go.mod
module github.com/Work-Fort/Scope/go/frontend

go 1.25.0
```

**Step 4: Tidy the module**

Run: `cd go/frontend && go mod tidy`
Expected: clean exit (validates go.mod syntax and go directive)

**Step 5: Commit**

```bash
git add mise.toml go/frontend/go.mod
git commit -m "feat(frontend-sdk): add Go tooling, init module at go/frontend"
```

---

### Task 2: Restore frontend.go from git history

**Files:**
- Create: `go/frontend/frontend.go`

**Step 1: Restore the file from commit 437de6a**

Run: `git show 437de6a:pkg/frontend/frontend.go > go/frontend/frontend.go`

**Step 2: Update the package — no changes needed**

The file is `package frontend` with only stdlib imports (`encoding/json`, `io/fs`, `net/http`, `strings`). The module path changed but the package name stays the same. No edits required.

**Step 3: Verify it compiles**

Run: `cd go/frontend && go build ./...`
Expected: clean exit, no errors

**Step 4: Commit**

```bash
git add go/frontend/frontend.go
git commit -m "feat(frontend-sdk): restore Handler and Manifest from git history"
```

---

### Task 3: Restore frontend_test.go and fix import path

**Files:**
- Create: `go/frontend/frontend_test.go`

**Step 1: Restore the test file from commit 437de6a**

Run: `git show 437de6a:pkg/frontend/frontend_test.go > go/frontend/frontend_test.go`

**Step 2: Update the import path**

Change:
```go
"github.com/Work-Fort/Scope/pkg/frontend"
```

To:
```go
"github.com/Work-Fort/Scope/go/frontend"
```

**Step 3: Run tests**

Run: `cd go/frontend && go test -v ./...`
Expected: all 8 tests pass:
- `TestHealthProbe_OK`
- `TestHealthProbe_Unavailable`
- `TestCacheHeaders_Assets`
- `TestCacheHeaders_RemoteEntry`
- `TestCacheHeaders_OtherFiles`
- `TestFileNotFound`
- `TestHealthProbe_WithManifest`
- `TestHealthProbe_Unavailable_IncludesManifest`

**Step 4: Commit**

```bash
git add go/frontend/frontend_test.go
git commit -m "test(frontend-sdk): restore tests, update import path"
```

---

### Task 4: Add mise tasks for the Go frontend SDK

**Files:**
- Create: `.mise/tasks/sdk/go/build`
- Create: `.mise/tasks/sdk/go/test`

**Step 1: Create the build task**

```bash
#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#MISE description="Compile Go frontend SDK"
set -euo pipefail

cd go/frontend
go build ./...
```

**Step 2: Create the test task**

```bash
#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#MISE description="Run Go frontend SDK tests"
set -euo pipefail

cd go/frontend
go test -v ./...
```

**Step 3: Make them executable**

Run: `chmod +x .mise/tasks/sdk/go/build .mise/tasks/sdk/go/test`

**Step 4: Verify tasks appear and run**

Run: `mise tasks | grep sdk`
Expected: `sdk:go:build` and `sdk:go:test` listed

Run: `mise run sdk:go:test`
Expected: all 8 tests pass

**Step 5: Update the CI task to include the SDK**

The existing `.mise/tasks/ci` uses `depends=["lint", "test"]`. Add `sdk:go:test` to its dependencies:

Change `#MISE depends=["lint", "test"]` to `#MISE depends=["lint", "test", "sdk:go:test"]`

**Step 6: Commit**

```bash
git add .mise/tasks/sdk/go/build .mise/tasks/sdk/go/test .mise/tasks/ci
git commit -m "chore: add mise tasks for Go frontend SDK (sdk:go:build, sdk:go:test)"
```

---

### Task 5: Add GitHub Actions workflow for Go SDK release

**Files:**
- Create: `.github/workflows/release-sdk-go.yml`

**Step 1: Create the workflow**

Model it on Passport's `release-sdk-go.yml`. Key differences from Passport:
- Paths filter: `go/frontend/**`
- Tag prefix: `go/frontend/v`
- Test command: `cd go/frontend && go test -v ./...`
- Module name in release notes: `github.com/Work-Fort/Scope/go/frontend`

```yaml
# SPDX-License-Identifier: Apache-2.0

name: Release Go SDK

on:
  push:
    branches: [master]
    paths:
      - 'go/frontend/**'

permissions:
  contents: write

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: jdx/mise-action@v3
      - run: mise run sdk:go:test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: test
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: go/frontend/v
          paths: go/frontend/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Create GitHub Release
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'
    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^go/frontend/v||')
          cat > release-notes.md << EOF
          # Go Frontend SDK v${VERSION}

          Module: \`github.com/Work-Fort/Scope/go/frontend\`

          ## What's Changed

          ${CHANGELOG}

          ## Usage

          \`\`\`bash
          go get github.com/Work-Fort/Scope/go/frontend@${TAG}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: Go SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 2: Commit**

```bash
git add .github/workflows/release-sdk-go.yml
git commit -m "ci: add Go frontend SDK release workflow"
```

---

### Task 6: Update documentation to reference new import path

**Files:**
- Modify: `docs/frontend/service-contract.md:65`
- Modify: `docs/frontend/architecture.md:7,32`
- Modify: `docs/frontend/getting-started/solidjs.md:145,157`
- Modify: `docs/frontend-docs-design.md:38,58,93`

**Step 1: Update `docs/frontend/service-contract.md`**

Change the code comment path:
```
// pkg/frontend/frontend.go  →  // go/frontend/frontend.go
```

**Step 2: Update `docs/frontend/architecture.md`**

Line 7: Change `pkg/frontend.Handler` to `go/frontend.Handler`
Line 32: Change `pkg/frontend/frontend.go` to `go/frontend/frontend.go` (both occurrences on that line)

**Step 3: Update `docs/frontend/getting-started/solidjs.md`**

Line 145: Change `pkg/frontend.Handler` to `go/frontend.Handler`
Line 157: Change the import path:
```go
"github.com/Work-Fort/Scope/pkg/frontend"  →  "github.com/Work-Fort/Scope/go/frontend"
```

**Step 4: Update `docs/frontend-docs-design.md`**

Line 38: Change `pkg/frontend.Handler` to `go/frontend.Handler`
Line 58: Change `pkg/frontend` to `go/frontend`
Line 93: Change `pkg/frontend.Handler` to `go/frontend.Handler`

**Step 5: Verify no live docs reference the old path**

Run: `grep -r "pkg/frontend" docs/ --include="*.md" | grep -v "archive/" | grep -v "plans/"`
Expected: no results

**Step 6: Commit**

```bash
git add docs/frontend/service-contract.md docs/frontend/architecture.md docs/frontend/getting-started/solidjs.md docs/frontend-docs-design.md
git commit -m "docs: update frontend SDK import path to go/frontend"
```

---

## Verification Checklist

- [ ] `mise install` installs Go 1.25
- [ ] `cd go/frontend && go build ./...` compiles cleanly
- [ ] `mise run sdk:go:build` exits cleanly
- [ ] `mise run sdk:go:test` passes all 8 tests
- [ ] `mise run ci` runs and includes Go SDK tests
- [ ] `grep -r "pkg/frontend" docs/ --include="*.md" | grep -v "archive/" | grep -v "plans/"` returns no results
- [ ] Workflow YAML is valid (review `.github/workflows/release-sdk-go.yml`)

---

## QA Testing

**Test scenarios:**
- Import the module from an external Go project: `go get github.com/Work-Fort/Scope/go/frontend@<commit>` resolves and compiles
- Health endpoint returns 200 + manifest JSON when `remoteEntry.js` exists in the FS
- Health endpoint returns 503 + manifest JSON when `remoteEntry.js` does not exist
- `/ui/assets/*` responses have `Cache-Control: public, max-age=31536000, immutable`
- `/ui/remoteEntry.js` and other `/ui/*` responses have `Cache-Control: no-cache`
- 404 responses have no `Cache-Control` header

**API/CLI commands to test:**
- `mise run sdk:go:test` — all 8 tests pass
- `mise run sdk:go:build` — exits cleanly
- `mise run ci` — includes Go SDK tests in the pipeline

**Success criteria:**
- Other WorkFort Go services can `go get` the module and use `frontend.Handler` to serve their federated UI
- CI workflow triggers on `go/frontend/**` changes and creates tagged releases
- All developer-facing docs reference `go/frontend`, not `pkg/frontend`
