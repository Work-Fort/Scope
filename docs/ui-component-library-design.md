# UI Component Library — Design Spec

## Goal

Expand `@workfort/ui` from a shell-chrome component set (14 components) into a full app development toolkit. Two audiences: human developers who want control, and AI agents that need predictable, composable primitives to build frontends quickly.

## Architecture

Two layers:

**Lion behavior layer** — `@lion/ui` (MIT, Lit-based) provides form validation, overlay management, and accessibility internals. Lion components handle keyboard navigation, ARIA attributes, focus trapping, and form participation. Lion is a `dependencies` entry in `@workfort/ui` — consumers never import from `@lion/ui` directly.

**WorkFort visual layer** — `WfElement` (extends LitElement, light DOM) renders markup and styles. Each `wf-*` component extends the corresponding Lion mixin/class where one exists, and extends `WfElement` directly where it doesn't (button, badge, divider, etc.).

Integration pattern for Lion-backed components:

```
Lion mixin/class (behavior, a11y, validation)
  → WfElement (light DOM, token-based styles)
    → wf-* custom element registration
```

Lion is used selectively — only where it provides behavioral complexity (forms, overlays, validation). Purely visual components stay as plain `WfElement` implementations.

**Why Lion over Shoelace fork:** The research recommended forking Shoelace as the fastest path. We chose Lion instead because: (1) Shoelace is frozen — owning a fork means maintaining 58 components of someone else's visual opinions, bug fixes, and accessibility debt; (2) Lion is actively maintained by ING Bank and designed explicitly as a foundation layer; (3) Lion gives us full visual control from day one, which matters for a platform where AI agents and human developers both need a distinctive, consistent identity — not a reskinned third-party library.

**Light DOM integration:** `WfElement` overrides `createRenderRoot()` to return `this` (light DOM). Lion's form mixins (`LionField`, `FormControlMixin`, `ValidateMixin`) operate on `_inputNode` references and template slots, not on `this.shadowRoot` directly. The pattern is to extend the Lion class, override `createRenderRoot()`, and provide our own `render()` method that places the `_inputNode` in light DOM. Lion's overlay system (`OverlayMixin`) creates its own container element appended to `document.body` — it does not depend on the host component's shadow root. This needs validation during Phase 2's first component implementation; if a specific Lion mixin requires shadow DOM, we use composition (wrapping) instead of inheritance.

## Design Tokens

CSS custom properties namespaced `--wf-*`, organized in three tiers. No build tooling needed — CSS is the source of truth.

**Files** in `web/packages/ui/src/styles/`:

| File | Tier | Contents |
|------|------|----------|
| `primitives.css` | 1 | Raw scales: stone palette (50–950), neutral gray, semantic color scales (red, amber, green, blue), spacing scale, font sizes, font weights, line heights, border radii, shadows, z-index layers, motion (duration + easing) |
| `tokens.css` | 2 | Semantic aliases referencing primitives: `--wf-color-bg` → `var(--wf-stone-50)`, `--wf-color-error` → `var(--wf-red-500)`. Dark mode via `[data-theme="dark"]` swaps which primitives semantic tokens point at |
| `components.css` | 3 | Component-level tokens and base styles: `--wf-button-radius`, `--wf-input-height`, `--wf-panel-padding`. Reference semantic tokens |

**Current state:** `tokens.css` hardcodes hex values. The refactor adds `primitives.css` and makes `tokens.css` reference it via `var()`.

**UnoCSS bridge:** The shell's `uno.config.ts` theme points at the same CSS vars (`colors: { bg: 'var(--wf-color-bg)' }`). One source of truth, two consumption paths (components and utility classes). No duplication.

## Token Categories

| Category | Primitive examples | Semantic examples |
|----------|-------------------|-------------------|
| Color | `--wf-stone-50` through `--wf-stone-950`, `--wf-red-500` | `--wf-color-bg`, `--wf-color-text`, `--wf-color-error`, `--wf-color-success` |
| Spacing | `--wf-space-1` through `--wf-space-96` (rem) | — (primitives used directly) |
| Typography | `--wf-text-xs` through `--wf-text-4xl`, `--wf-font-sans`, `--wf-weight-normal` | `--wf-font-body`, `--wf-font-heading` |
| Radius | `--wf-radius-sm`, `--wf-radius-md`, `--wf-radius-lg`, `--wf-radius-full` | — |
| Shadow | `--wf-shadow-sm`, `--wf-shadow-md`, `--wf-shadow-lg` | — |
| Z-index | `--wf-z-dropdown`, `--wf-z-modal`, `--wf-z-toast`, `--wf-z-tooltip` | — |
| Motion | `--wf-duration-fast`, `--wf-duration-normal`, `--wf-ease-in-out` | — |

## Component Inventory

### Existing (14) — keep, refactor to use tokens

Panel, Button, Badge, StatusDot, Skeleton, Divider, TextInput, List, ListItem, ScrollArea, ErrorFallback, Banner, Toast, ToastContainer.

`wf-text-input` is replaced by Lion-backed `wf-input` in Phase 2.

### Phase 2: Forms (Lion-backed)

Input, Textarea, Select, Combobox, Checkbox, Radio, Toggle/Switch, Slider, Date Picker, File Upload, Form (validation + submission).

Lion provides: `form-core` (validation lifecycle, dirty/touched/submitted state), `validate-messages` (localized errors), overlay management (for combobox/select dropdowns), `fieldset` (grouped validation).

### Phase 3: Layout & Display

Dialog/Modal, Drawer, Tabs, Accordion, Card, Tooltip, Popover, Table/Data Grid.

Lion provides: overlay primitives (for dialog, drawer, tooltip, popover). Table/Data Grid is built from scratch.

### Phase 4: Navigation & Feedback

Breadcrumbs, Pagination, Stepper, Progress Bar, Spinner, Alert/Confirm Dialog.

Lion provides: pagination, progress-indicator, steps.

## API Conventions

Across all components:

- Props: kebab-case attributes, camelCase properties
- Events: `wf-` prefix (`wf-change`, `wf-submit`, `wf-click`)
- Variants: string attribute (`variant="filled"`)
- Form components: consistent `disabled`, `readonly`, `required` attributes
- TypeScript types exported for every component class
- React wrappers generated for all new components (extending the existing `wrapWc` pattern)

## Phasing

| Phase | Scope | Unblocks |
|-------|-------|----------|
| 1 — Design Tokens | `primitives.css`, refactor `tokens.css`, UnoCSS bridge, migrate existing 14 components to token refs | Sharkfin/Passport/Hive can build with consistent theming |
| 2 — Forms | Lion integration, 11 form components, form validation | Real app development with data entry |
| 3 — Layout & Display | 8 components including data grid | Admin panels, settings pages, data-heavy UIs |
| 4 — Navigation & Feedback | 6 components | Complete toolkit for AI agents to compose full UIs |

Each phase gets its own implementation plan. Phase 1 is delivered first.

## Framework Adapter Updates

Each phase that adds components also updates the framework adapters:

- `@workfort/ui-react` — add React wrappers for new components (using existing `wrapWc` factory)
- `@workfort/ui-solid`, `@workfort/ui-vue`, `@workfort/ui-svelte` — no component wrappers needed (these frameworks handle custom elements natively). Phase 2 adds form validation composables/stores for each adapter.

## Testing

Each phase includes:

- Unit tests (Vitest + happy-dom) for component behavior and accessibility
- axe-core integration in component tests for automated a11y audits
- All existing tests must continue to pass after token migration (Phase 1)

## Versioning & Publishing

Phase 1 (tokens) is a **minor version bump** — no public API changes, only internal CSS refactoring. Existing class names and attributes are preserved.

Phase 2 (Lion integration) is a **major version bump** — adds `@lion/ui` as a dependency, increases bundle size, introduces new components. Lion is tree-shakeable: consumers who only use existing components (panel, button, etc.) should not pay for Lion's form system in their bundle.

Publishing uses the existing per-package GitHub Actions workflows with OIDC npm auth.

## Consumer Migration

Phase 1: No breaking changes. Existing components render identically — only the internal CSS implementation changes from hardcoded hex to `var()` references. Token names that change are documented in a changelog.

Phase 2: `wf-text-input` is deprecated in favor of `wf-input`. Both work during the transition. A migration note documents the mapping.

## Cross-Phase Dependencies

Lion's overlay system is first used in Phase 2 (select/combobox dropdowns) and reused in Phase 3 (dialog, drawer, tooltip, popover). Phase 2's implementation plan must include the overlay foundation work so Phase 3 can build on it.

## Research References

- `docs/web-component-library-research.md` — full library evaluation
- `docs/design-token-research.md` — token system analysis
