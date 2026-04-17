# Remaining Work — Scope

Tracks known bugs and follow-ups. Items are roughly priority-ordered within each section.

---

## Open

### Bugs
- [ ] **Local-fort discovery ignores Pylon** — `scope-server/src/main.rs` branches the discovery polling loop as `if fort.local { probe_all } else { fetch_from_pylon }`. For local forts that also have `pylon:` set, only the static `services:` list is probed and Pylon is never queried. The intent (per the integration-env doc) is that Passport is listed directly for pre-auth bootstrap but the rest of the service catalog comes from Pylon. Fix: when `local && pylon is Some`, fetch from Pylon and merge results with the direct probes of `fort.services`. Workaround today: list all service URLs explicitly in `~/.config/workfort/config.yaml`.
