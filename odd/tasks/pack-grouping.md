# Pack grouping

## Objective
Group catalog skills by install pack (parent + children), the way skills.sh groups by GitHub repo.

## Problem
Installed packs flatten into sibling folders. `bug-bounty` does not prefix-match `hunt-*`, so the UI never shows a parent with its children.

## Why
Users install a repo of many skills and need to browse that pack as one family.

## Scope
- Infer pack from: alias (parent present), nested directory under `/skills/`, shared `sourcePkg`, then name prefix.
- Catalog grid grouped by pack; categories stay a secondary filter.
- Out of scope this slice: description-based taxonomy cleanup, search tokenization.

## TDD
- Mode: on (session Strict TDD)
- Runner: `cd backend && go test ./internal/taxonomy/ ./internal/service/`

## Tasks
- [x] T1 Infer pack families (alias, directory, sourcePkg) with tests
- [x] T2 Catalog UI grouped by pack
- [x] T3 Work-unit commits — `31b6240` feat(catalog): group skills by install pack

## Verification
- `cd backend && go test ./internal/taxonomy/ ./internal/service/`: pass
- `cd backend && go test ./...`: 62 passed
- `cd frontend && bun run build`: pass
- Live scan: family `bug-bounty` = 82 unique, parent `bug-bounty` childCount 81; family `hunt` leftover 0; family `erp` = 19
- Browser UI: not exercised (no browser session); grouping is in OrganizerApp.vue catalog grid

## Acceptance
- `bug-bounty` present → `hunt-*` and `bb-*` share family `bug-bounty`, parent is `bug-bounty`
- `hunt-*` only → family stays `hunt`
- Hermes `skills/erp/vue/SKILL.md` → family `erp`
- Shared `sourcePkg` → family = repo slug
