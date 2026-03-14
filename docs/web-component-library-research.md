# Web Component Library Research

**Date:** 2026-03-14
**Purpose:** Evaluate existing open-source web component libraries as a potential foundation for the WorkFort design system.

---

## Requirements Recap

- Web components (custom elements), not framework-specific
- Permissive license: MIT, Apache-2.0, or BSD
- Actively maintained (commits within 6 months of March 2026)
- Comprehensive coverage: forms, data display, layout, navigation, feedback
- Works with Lit (our components extend LitElement)
- Styleable, predictable APIs (consistent for AI agent use)

---

## Library Evaluations

### 1. Shoelace (now superseded by Web Awesome)

| Field | Detail |
|-------|--------|
| **License** | MIT |
| **Stars** | 13,874 |
| **Last push** | March 5, 2026 |
| **Actively maintained?** | **NO** — explicitly states "Shoelace is no longer actively being developed." The codebase has been migrated into the commercial Web Awesome product by Font Awesome. |

**Components (~58):** alert, animated-image, animation, avatar, badge, breadcrumb, button, button-group, card, carousel, checkbox, color-picker, copy-button, details, dialog, divider, drawer, dropdown, format-bytes, format-date, format-number, icon, icon-button, image-comparer, input, menu, menu-item, option, popup, progress-bar, progress-ring, qr-code, radio, radio-group, range, rating, relative-time, select, skeleton, spinner, split-panel, switch, tab, tab-group, tag, textarea, tooltip, tree, tree-item + others. Excellent breadth.

**Uses Lit?** Yes — built on LitElement.
**Shadow DOM:** Full shadow DOM with well-documented `::part()` selectors.
**Theming:** CSS custom properties with `--sl-` prefix design tokens. Override at `:root`. CSS parts for per-component styling.
**Bundle size:** Roughly 80–100 KB minified+gzipped for the full bundle; tree-shakable per-component imports.
**API quality:** Excellent. Consistent `variant`, `size`, `disabled` props across all components. Clean events with `sl-` prefix. Good TypeScript types.
**Wrappable/Extendable:** Yes, all classes are exported. Can subclass any component.
**Dealbreakers:** **Not maintained.** The successor (Web Awesome) has a free tier but is a commercial product with paid Pro components. Its open-source status is unclear. Using a dead library as a foundation is a risk.

---

### 2. Lion (by ING Bank)

| Field | Detail |
|-------|--------|
| **License** | MIT |
| **Stars** | 1,936 |
| **Last push** | March 14, 2026 |
| **Actively maintained?** | Yes — active daily commits |

**Components (~40 in `@lion/ui`):** accordion, button, calendar, checkbox-group, collapsible, combobox, core, dialog, drawer, fieldset, form-core, form-integrations, form, helpers, icon, input, input-amount, input-date, input-datepicker, input-email, input-file, input-iban, input-range, input-stepper, input-tel, input-tel-dropdown, listbox, overlays, pagination, progress-indicator, radio-group, select, select-rich, steps, switch, tabs, textarea, tooltip, validate-messages. Strong form coverage; limited data display (no table/grid).

**Uses Lit?** Yes — LitElement is the base class.
**Shadow DOM:** Primarily shadow DOM, with a "white label" philosophy — **zero default visual styles**. This is intentional; Lion is meant to be a bare accessible foundation.
**Theming:** No opinionated tokens. You style from scratch using your own CSS. This is both a strength (total control) and a weakness (you must build everything visually yourself).
**Bundle size:** Very small individual packages, ~3–8 KB per component gzipped.
**API quality:** Solid, consistent form-centric API. Good TypeScript. Heavy focus on form validation and overlays system.
**Wrappable/Extendable:** This is explicitly the intended use case — Lion is designed to be extended, not used directly.
**Dealbreakers:** No visual styling at all. For a design system foundation, this means no visual scaffold to build on — you'd be doing all visual work from scratch. No data grid or table component. The "new portal" note on their site suggests some maintenance concerns.

---

### 3. Spectrum Web Components (by Adobe)

| Field | Detail |
|-------|--------|
| **License** | Apache-2.0 |
| **Stars** | 1,493 |
| **Last push** | March 13, 2026 |
| **Actively maintained?** | Yes — daily commits |

**Components (~96 packages):** Accordion, Action Bar, Action Button, Action Group, Action Menu, Alert Banner, Alert Dialog, Asset, Avatar, Badge, Breadcrumbs, Button, Card, Checkbox, Color Area/Field/Slider/Wheel, Coachmark, Dialog, Divider, Dropzone, Field Group, Field Label, Help Text, Icon, Illustrated Message, Link, Menu, Number Field, Overlay, Picker, Popover, Progress Bar/Circle, Radio, Search, Sidenav, Slider, Status Light, Swatch, Switch, Table, Tabs, Tag, Textfield, Thumbnail, Toast, Tooltip, Tray, Underlay, and more. Exceptional breadth including a data table.

**Uses Lit?** Yes — LitElement base class explicitly called out in their docs.
**Shadow DOM:** Full shadow DOM, but styled via Spectrum's design token system.
**Theming:** CSS custom properties from the `@spectrum-css` token system. Supports multiple Spectrum themes (light, dark, high-contrast). Token-based theming is thorough.
**Bundle size:** Modular, per-package. Individual components ~5–15 KB gzipped.
**API quality:** Good. Consistent with Spectrum design system patterns. TypeScript types generated from their component models. Props follow Spectrum naming conventions (`variant`, `quiet`, `emphasized`).
**Wrappable/Extendable:** Technically yes, but the components are tightly coupled to Spectrum visual language — they look definitively like Adobe Spectrum. Overriding the visual identity while keeping behavior would require extensive token overrides and CSS parts work.
**Dealbreakers:** Strong visual identity coupling to Adobe Spectrum. The `sp-*` prefix and Spectrum aesthetic would be very hard to rebrand for a non-Adobe product. Theming requires adopting the full Spectrum token system. Not a blank canvas.

---

### 4. Vaadin Web Components

| Field | Detail |
|-------|--------|
| **License** | Apache-2.0 (confirmed via npm) |
| **Stars** | 556 |
| **Last push** | March 13, 2026 |
| **Actively maintained?** | Yes — daily commits |

**Components (~50+):** Checkbox, Combo Box, Custom Field, Date Picker, Date Time Picker, Email Field, List Box, Number Field, Message Input, Multi-Select Combo Box, Password Field, Radio Button, Rich Text Editor, Select, Text Area, Text Field, Time Picker, Upload, Auto Grid, Accordion, Avatar, Badge, Button, Charts, Card, Confirm Dialog, Context Menu, CRUD, Details, Dialog, Grid (with sorting/filtering/lazy load), Grid Pro, Icons, Markdown, Menu Bar, Message List, Notification, Popover, Progress Bar, Scroller, Split Layout, Tabs, Tooltip, Virtual List. Outstanding data grid; rich enterprise components.

**Uses Lit?** Yes — depends on `lit` in package.json.
**Shadow DOM:** Full shadow DOM with `vaadin-themable-mixin` for theming.
**Theming:** CSS custom properties via Lumo design system (default theme). Theme tokens can be overridden. Supports Lumo and Material themes. Has an official Figma library.
**Bundle size:** Per-component packages; Grid is heavier (~50 KB gzipped).
**API quality:** Good, enterprise-grade API. Consistent patterns across form components. Strong TypeScript. Some complexity in data binding patterns for Grid (may be verbose for AI agents).
**Wrappable/Extendable:** Yes, the `vaadin-themable-mixin` is designed for theming. The Lumo tokens can be fully overridden.
**Dealbreakers:** The Lumo design system aesthetic is distinctly "Vaadin." The 943 open issues indicates a large backlog. Components are enterprise-oriented which may be over-engineered for simpler use cases. The open-source components are a subset of their paid Vaadin platform.

---

### 5. Microsoft FAST

| Field | Detail |
|-------|--------|
| **License** | MIT |
| **Stars** | 9,636 |
| **Last push** | March 13, 2026 |
| **Actively maintained?** | Yes |

**Important:** FAST is a **web component framework/runtime**, not a ready-to-use component library. It provides `FASTElement` (a base class similar to `LitElement`), a template engine, and reactive property system. The actual components built on FAST are in **Fluent UI Web Components** (separate repo). FAST packages: `fast-element`, `fast-html`, `fast-router`, `fast-ssr`.

**Uses Lit?** No — FAST is an alternative to Lit. They are competing web component frameworks.
**Dealbreaker:** Not Lit-compatible. Cannot be used alongside our existing LitElement components without running two separate web component frameworks. Not a component library.

---

### 6. PatternFly Elements (by Red Hat)

| Field | Detail |
|-------|--------|
| **License** | MIT |
| **Stars** | 389 |
| **Last push** | March 8, 2026 |
| **Actively maintained?** | Yes |

**Components (~34):** Accordion, Alert, Avatar, Back To Top, Background Image, Badge, Banner, Button, Card, Chip, Clipboard Copy, Code Block, Dropdown, Helper Text, Hint, Icon, Jump Links, Label, Modal, Panel, Popover, Progress, Progress Stepper, Search Input, Select, Spinner, Switch, Table, Tabs, Text Area, Text Input, Tile, Timestamp, Tooltip.

**Uses Lit?** Yes — explicitly depends on `lit`, built on LitElement.
**Shadow DOM:** Full shadow DOM.
**Theming:** CSS custom properties with `--pf-` prefix tokens. Aligns with PatternFly's token system.
**Bundle size:** Lightweight — 3–10 KB per component gzipped.
**API quality:** Consistent, uses `pf-` element prefix. TypeScript types. Moderate complexity.
**Wrappable/Extendable:** Yes, exported classes.
**Dealbreakers:** Red Hat / RHEL enterprise aesthetic — very "enterprise CRUD app" look. Smaller community (389 stars). The component set is missing some key pieces (no date picker, no combobox, no data grid). PatternFly visual identity is distinct and may be hard to rebrand.

---

### 7. Carbon Web Components (by IBM)

| Field | Detail |
|-------|--------|
| **License** | Apache-2.0 |
| **Stars** | 472 |
| **Last push** | March 8, 2023 |
| **Actively maintained?** | **NO** — **archived in March 2023**. |

**Dealbreaker:** Archived. IBM migrated to `@carbon/web-components` within the main Carbon monorepo (`carbon-design-system/carbon`). The original repo is dead. The new location exists but the web components portion is secondary to React, and the commit history shows it's not a standalone focus.

---

### 8. Elix

| Field | Detail |
|-------|--------|
| **License** | MIT |
| **Stars** | 834 |
| **Last push** | March 2023 (last real feature commit: April 2022) |
| **Actively maintained?** | **NO** — effectively abandoned. Last meaningful commit was April 2022. |

**Dealbreaker:** Not maintained. Uses vanilla web components (no Lit). Small niche library focused on UI patterns (carousels, lists, menus) rather than a comprehensive component system.

---

## Additional Library: Material Web (by Google)

Discovered during research — not in the original list but highly relevant.

| Field | Detail |
|-------|--------|
| **License** | Apache-2.0 |
| **Stars** | 10,798 |
| **Last push** | March 14, 2026 |
| **Actively maintained?** | Yes — multiple commits per week |

**Components (~23):** Button, Checkbox, Chips, Color (utilities), Dialog, Divider, Elevation, FAB, Field, Focus Ring, Icon, Icon Button, List, Menu, Progress (linear/circular), Radio, Ripple, Select, Slider, Switch, Tabs, Textfield, Typography.

**Uses Lit?** Yes — explicitly built on LitElement (confirmed in source: `import {html, LitElement} from 'lit'`).
**Shadow DOM:** Full shadow DOM.
**Theming:** CSS custom properties via Material Design 3 tokens (`--md-` prefix). The M3 token system is very well-specified and composable.
**Bundle size:** Modular. Individual components 5–15 KB gzipped.
**API quality:** Good, clean, minimal API surface. Consistent `variant` and `label` patterns. TypeScript types. Relatively simple prop surfaces — good for AI agents.
**Wrappable/Extendable:** Yes, classes are exported. However, the Material Design 3 visual identity is strongly opinionated (Google/Android aesthetic).
**Dealbreakers:** Material Design visual identity — if you don't want MD3 aesthetics, rebranding requires overriding the entire `--md-*` token system. No data table component. Limited component count compared to Shoelace or Vaadin.

---

## Comparison Summary

| Library | License | Lit? | Active? | Components | Theming | Best For |
|---------|---------|------|---------|------------|---------|----------|
| **Shoelace** | MIT | Yes | NO (archived) | ~58 | CSS tokens + parts | — |
| **Lion** | MIT | Yes | Yes | ~40 | None (white label) | Accessible forms foundation |
| **Spectrum WC** | Apache-2.0 | Yes | Yes | ~96 | Spectrum tokens | Adobe Spectrum apps |
| **Vaadin** | Apache-2.0 | Yes | Yes | ~50+ | Lumo tokens | Enterprise / data-heavy apps |
| **FAST** | MIT | NO | Yes | 0 (framework) | — | Building WC frameworks |
| **PatternFly** | MIT | Yes | Yes | ~34 | PF tokens | Red Hat / enterprise apps |
| **Carbon WC** | Apache-2.0 | ? | NO (archived) | ~? | — | — |
| **Elix** | MIT | No | NO | ~20 | None | — |
| **Material Web** | Apache-2.0 | Yes | Yes | ~23 | MD3 tokens | Material Design apps |

---

## Recommendation

### Verdict: Build on Shoelace's codebase (fork) or build from scratch extending our Lit components.

**None of the actively maintained libraries are ideal for a clean custom design system.** Here is the reasoning:

**What makes this hard:**
- Every actively-maintained library with good breadth (Spectrum, Vaadin, PatternFly, Material Web) carries a strong visual identity that is difficult to override cleanly for a custom product.
- The one library that was designed explicitly to be extended without visual opinions (Lion) has zero visual defaults — meaning you'd be doing the full visual design layer yourself anyway.
- Shoelace was the perfect match (MIT, Lit, 58 components, CSS parts theming, clean API) but it is **no longer maintained** — the author migrated to a commercial product.

**The realistic options ranked:**

**Option A: Fork Shoelace (recommended if using a third-party base)**

Shoelace is MIT, built on Lit, has ~58 components with the best API design of the group, uses CSS custom properties + `::part()` theming that maps directly to a custom design token system, and every component class is exportable. It's frozen in time as of late 2025 (v2.20.1) but it's complete enough to use as a foundation. A fork lets us own future development, apply our own CSS token layer, and extend with `wf-` prefixed components. The main risk is owning all future bug fixes and accessibility updates ourselves.

**Option B: Lion as an accessible behavior layer + custom visuals on top**

Lion's explicitly white-label architecture means your design system sits completely on top — no visual debt. You get Lion's excellent form validation system, overlay management, and accessibility primitives. This is the architecture ING itself uses to build their own branded components. However, it requires building all visual components from scratch, which is equivalent to Option C.

**Option C: Build from scratch, extending our existing LitElement base**

Given that we already have a Lit base class and component infrastructure, and given that any library we choose requires substantial CSS override work to rebrand, the delta between "fork Shoelace" and "build our own" may be smaller than it appears. Building from scratch gives total API control — which matters for AI agent usability — and zero dependency risk. The cost is time to build the component set.

**Recommendation: Fork Shoelace (Option A) as a starting point.**

Rationale:
1. The API design is the best of the group — consistent, predictable, well-documented. This is the highest-value thing to inherit.
2. 58 components with solid accessibility already implemented saves months of work.
3. MIT license is clean for Apache-2.0 project (MIT is compatible with Apache-2.0).
4. Lit base matches our existing architecture exactly. No framework mismatch.
5. CSS custom properties + `::part()` approach is exactly what a design token system needs.
6. Freezing on a known-good version is acceptable if we own the fork. Security and bug fixes become our responsibility, which is already true of any component we build.

**If the team is unwilling to own a fork:** Use Lion as the form/behavior foundation + build the visual layer ourselves (Option B). This avoids owning Shoelace's full codebase while still getting Lion's excellent accessibility and form validation primitives.

**Discard:** FAST (wrong framework), Elix (abandoned), Carbon WC (archived), Spectrum (too Adobe-branded), Material Web (too Google-branded), Vaadin (enterprise/Lumo-branded).
