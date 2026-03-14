# Design Token System Research

*Research date: 2026-03-14*

---

## 1. Standards & Tools Survey

### DTCG (Design Tokens Community Group) Spec

- **Status**: Preview draft (2025.10, published 2026-03-09). The spec says explicitly: "Do not attempt to implement this version of the specification. Do not reference this version as authoritative in any way."
- **Verdict for adoption**: NOT ready. The format has been in flux for years. The token type structure changed significantly in the 2025.10 draft (e.g., `dimension` is now `{ value: number, unit: "px"|"rem" }` not a string like `"8px"`). Tooling is catching up. Watch but do not adopt as source of truth yet.
- **What IS stable from DTCG**: The three-tier mental model (primitive → semantic → component), the `$value`/`$type`/`$description` property conventions, and the alias/reference concept (`{color.stone.900}`). These are safe to follow.

### Style Dictionary (Amazon)

- **License**: Apache 2.0
- **What it does**: Build tool. Reads token files (JSON/DTCG format), applies transforms, outputs platform-specific artifacts: CSS custom properties, JS ES modules, iOS Swift, Android XML, etc.
- **Version**: v4 is current (breaking change from v3). Forward-compatible with DTCG.
- **Verdict**: The right tool if you need multi-platform output (CSS + JS tokens + potentially native). For a pure web project it's potentially overkill — you'd be adding a build step that outputs what you could write directly.

### Theo (Salesforce)

- **Status**: Effectively abandoned/unmaintained. The npm package exists but the repo shows minimal activity. Not recommended.

### Open Props

- **License**: MIT (Adam Argyle, 2021)
- **What it is**: Pre-built CSS custom properties — NOT a token build tool. It's ~500 props covering: sizes (fluid + fixed), colors (full palettes), typography, borders, shadows, animations, easings, gradients, z-index, aspect ratios.
- **Approach**: Opinionated primitives you consume directly or override. No semantic layer. Import individual prop-packs (`open-props/sizes`, `open-props/easings`) or the full bundle (4.0kB min).
- **Verdict**: Useful as a *reference* and source of inspiration for naming conventions and the animation/easing category. Don't adopt it wholesale — it uses generic numeric names (`--size-3`, `--radius-2`) that conflict with the `--wf-*` namespace and semantic intentions.

### Tailwind's Token Approach

Tailwind organizes tokens across three concerns:
1. **Scales** — spacing (`0.5rem` steps), font sizes (`xs` through `9xl`), radii, shadows as named stops in a numeric/named scale.
2. **Color palettes** — full 50–950 numeric scales per hue. Semantic colors reference these (e.g., `red-500`).
3. **CSS variable bridge** — Tailwind v4 shifts to CSS `@theme` variables; UnoCSS supports `theme.colors` values as `var(--wf-*)` references.

The UnoCSS `theme` config already bridges CSS vars to utility classes. The existing `uno.config.ts` does this correctly.

---

## 2. Token Tier Model

Every mature design system (Carbon, Shoelace/Web Awesome, Spectrum, Primer) converges on three tiers:

### Tier 1: Primitive Tokens (Scale)
Raw values. No semantic meaning. Never used directly in components.
```css
/* Examples */
--wf-stone-100: #f5f5f4;
--wf-stone-900: #1c1917;
--wf-space-2: 8px;
--wf-radius-2: 6px;
```

### Tier 2: Semantic Tokens (Aliases)
Named by *intent*, not value. Reference primitives. These are what components consume.
```css
/* Examples */
--wf-color-bg: var(--wf-stone-950);
--wf-color-text: var(--wf-stone-100);
--wf-color-border: var(--wf-stone-800);
--wf-color-primary: var(--wf-stone-100);
--wf-color-error: #ef4444;
--wf-space-component-gap: var(--wf-space-3);
```

### Tier 3: Component Tokens
Per-component overrides. Allow theming a single component without touching the semantic layer.
```css
/* Examples */
--wf-button-height-sm: 28px;
--wf-button-height-md: 36px;
--wf-input-height: 36px;
--wf-input-border-radius: var(--wf-radius-md);
```

---

## 3. What Tokens Are Needed (Tier 3 Library)

### Color
- **Primitive scale**: stone-50 through stone-950 (plus any brand/accent hues)
- **Semantic - surfaces**: `bg`, `bg-secondary`, `bg-elevated`, `bg-overlay`
- **Semantic - text**: `text-primary`, `text-secondary`, `text-muted`, `text-on-accent`, `text-disabled`
- **Semantic - border**: `border`, `border-focus`, `border-strong`
- **Semantic - interactive states**: `-hover`, `-active`, `-disabled` variants for each surface
- **Semantic - status**: `error`, `error-subtle`, `warning`, `warning-subtle`, `success`, `success-subtle`, `info`, `info-subtle`

### Typography
- `font-sans`, `font-mono`
- `font-size-xs` (0.6875rem/11px), `font-size-sm` (0.75rem/12px), `font-size-base` (0.875rem/14px), `font-size-md` (1rem), `font-size-lg` (1.125rem), `font-size-xl` (1.25rem)
- `font-weight-regular` (400), `font-weight-medium` (500), `font-weight-semibold` (600), `font-weight-bold` (700)
- `line-height-tight` (1.25), `line-height-normal` (1.5), `line-height-relaxed` (1.75)
- `letter-spacing-tight`, `letter-spacing-normal`, `letter-spacing-wide`

### Spacing
8-point grid base, half-point for tight cases:
- `space-1` = 4px, `space-2` = 8px, `space-3` = 12px, `space-4` = 16px, `space-5` = 20px, `space-6` = 24px, `space-8` = 32px, `space-10` = 40px, `space-12` = 48px, `space-16` = 64px

### Border Radius
- `radius-none` = 0, `radius-xs` = 3px, `radius-sm` = 6px, `radius-md` = 8px, `radius-lg` = 12px, `radius-xl` = 16px, `radius-full` = 9999px

### Shadows / Elevation
- `shadow-none`, `shadow-sm` (subtle border replacement), `shadow-md` (dropdowns), `shadow-lg` (modals), `shadow-xl` (toasts/tooltips)

### Motion
- `duration-instant` = 0ms, `duration-fast` = 100ms, `duration-normal` = 200ms, `duration-slow` = 300ms, `duration-deliberate` = 500ms
- `easing-linear`, `easing-standard` = `cubic-bezier(0.2, 0, 0, 1)`, `easing-decelerate`, `easing-accelerate`, `easing-spring`
- Respect `prefers-reduced-motion` — set durations to 0ms or 1ms in reduced-motion context

### Breakpoints
- `breakpoint-sm` = 640px, `breakpoint-md` = 768px, `breakpoint-lg` = 1024px, `breakpoint-xl` = 1280px, `breakpoint-2xl` = 1536px

### Z-Index
- `z-below` = -1, `z-base` = 0, `z-raised` = 10, `z-dropdown` = 100, `z-sticky` = 200, `z-overlay` = 300, `z-modal` = 400, `z-toast` = 500, `z-tooltip` = 600

### Component Tokens (examples)
```css
/* Form controls */
--wf-input-height-sm: 28px;
--wf-input-height-md: 36px;
--wf-input-height-lg: 44px;
--wf-input-padding-x: var(--wf-space-3);
--wf-input-font-size: var(--wf-font-size-sm);
--wf-input-border-radius: var(--wf-radius-sm);

/* Buttons */
--wf-button-height-sm: 28px;
--wf-button-height-md: 32px;
--wf-button-height-lg: 40px;
--wf-button-padding-x-sm: var(--wf-space-3);
--wf-button-padding-x-md: var(--wf-space-4);

/* Focus */
--wf-focus-ring-width: 2px;
--wf-focus-ring-offset: 2px;
--wf-focus-ring-color: var(--wf-color-border-focus);
```

---

## 4. How Existing Design Systems Structure Tokens

### Shoelace (now Web Awesome)
- Two-tier: primitives (via `--sl-color-*` numeric scales) + component tokens (`--sl-button-height-*`, `--sl-input-border-*`)
- Component tokens reference semantic tokens, semantic tokens reference primitives
- All defined as CSS custom properties on `:root`
- Light and dark themes in separate CSS files (`light.css`, `dark.css`)

### Carbon (IBM)
- Three-tier: **core tokens** (semantic, e.g., `$background`, `$layer-01`, `$text-primary`) + **component tokens** (per-component, never shared)
- Core tokens vary by theme (White, Gray 10, Gray 90, Gray 100)
- Component tokens (`$button-primary`, `$tag-background-*`) are isolated — not for cross-component use
- State tokens explicitly: `$background-hover`, `$background-active`, `$background-selected`

### Spectrum (Adobe)
- Four-tier: Global → Alias → Component → Variant
- Explicitly versioned tokens across component generations
- Heavily codified, designed for enterprise scale

---

## 5. Recommendations

### 5.1 Token Format

**Use CSS custom properties as the single source of truth.** Do not adopt DTCG JSON format yet — the spec is still in preview and requires a build step. For a small-to-medium library, CSS custom properties are:
- Framework-agnostic by definition
- Directly consumable by web components (light DOM), Solid, React, Vue, Svelte
- Inspectable in browser devtools without a build step
- Themeable at runtime (dark/light toggle works without JS)

If you later need multi-platform output (e.g., React Native, iOS), add Style Dictionary at that point. The migration path is straightforward: your CSS variables become the JSON token values.

### 5.2 Build Tooling

**No build tool needed initially.** Maintain `tokens.css` directly. When the token set grows or you need JS token exports (for use in Storybook, tests, or non-CSS consumers), add Style Dictionary v4. It can ingest your existing CSS and output JS ES modules alongside regenerated CSS.

Avoid Theo — effectively abandoned.

### 5.3 Token Organization Strategy

Adopt the three-tier structure:

```
packages/ui/src/styles/
  primitives.css      ← Tier 1: raw scales (stone palette, space numbers, etc.)
  tokens.css          ← Tier 2: semantic aliases (bg, text, border, status colors)
  components.css      ← Tier 3: component-specific tokens + component styles
```

**Current state**: `tokens.css` already IS semantic. The primitive scale (individual stone color stops) is missing — they're hardcoded as hex values directly in semantic tokens. This makes palette swapping impossible.

**Recommended refactor**:
1. Add `primitives.css` with the full stone scale + any other palette hues
2. Update `tokens.css` to reference primitives via `var(--wf-stone-*)`
3. Add missing token categories: typography weights, line-heights, motion, z-index, component tokens

### 5.4 Bridging to UnoCSS Theme Config

The existing `uno.config.ts` already uses the correct pattern — UnoCSS theme colors point to CSS vars:

```ts
colors: {
  wf: {
    bg: 'var(--wf-bg)',
    // ...
  }
}
```

**Extend this to cover all token categories**:

```ts
// uno.config.ts
export default defineConfig({
  presets: [presetWind()],
  theme: {
    colors: {
      wf: {
        bg: 'var(--wf-color-bg)',
        'bg-secondary': 'var(--wf-color-bg-secondary)',
        text: 'var(--wf-color-text)',
        'text-secondary': 'var(--wf-color-text-secondary)',
        'text-muted': 'var(--wf-color-text-muted)',
        border: 'var(--wf-color-border)',
        error: 'var(--wf-color-error)',
        warning: 'var(--wf-color-warning)',
        success: 'var(--wf-color-success)',
      },
    },
    spacing: {
      // Map semantic spacing names to vars OR just use Tailwind's
      // default scale + add wf-specific aliases
      'wf-xs': 'var(--wf-space-1)',
      'wf-sm': 'var(--wf-space-2)',
      'wf-md': 'var(--wf-space-4)',
      'wf-lg': 'var(--wf-space-6)',
      'wf-xl': 'var(--wf-space-8)',
    },
    borderRadius: {
      'wf-sm': 'var(--wf-radius-sm)',
      'wf-md': 'var(--wf-radius-md)',
      'wf-lg': 'var(--wf-radius-lg)',
    },
    fontFamily: {
      'wf-sans': 'var(--wf-font-sans)',
      'wf-mono': 'var(--wf-font-mono)',
    },
    fontSize: {
      'wf-xs': 'var(--wf-font-size-xs)',
      'wf-sm': 'var(--wf-font-size-sm)',
      'wf-base': 'var(--wf-font-size-base)',
      'wf-lg': 'var(--wf-font-size-lg)',
    },
    transitionDuration: {
      'wf-fast': 'var(--wf-duration-fast)',
      'wf-normal': 'var(--wf-duration-normal)',
      'wf-slow': 'var(--wf-duration-slow)',
    },
    zIndex: {
      'wf-dropdown': 'var(--wf-z-dropdown)',
      'wf-modal': 'var(--wf-z-modal)',
      'wf-toast': 'var(--wf-z-toast)',
      'wf-tooltip': 'var(--wf-z-tooltip)',
    },
  },
});
```

This means every token is defined once (in CSS), and the UnoCSS config is just a thin pointer layer. No duplication, no sync issues.

### 5.5 Open Props as Inspiration

Borrow Open Props' naming and values for motion tokens — it has excellent easing curves and animation presets. Do not import the library; copy the specific easing and duration values you want into `primitives.css`. Adam Argyle's choices for spring easings and reduced-motion handling are particularly good reference.

---

## 6. Gap Analysis: Current tokens.css vs Recommended

| Category | Current | Missing |
|---|---|---|
| Color – bg surfaces | bg, bg-secondary | bg-elevated, bg-overlay, bg-inverse |
| Color – text | text, text-secondary, text-muted | text-disabled, text-on-accent, text-inverse |
| Color – border | border | border-focus, border-strong, border-disabled |
| Color – interactive states | (none) | hover/active/disabled variants for surfaces |
| Color – status | error, warning, success | info, info-subtle |
| Color – primitives | (none — hardcoded hex) | Full stone 50–950 scale |
| Typography – sizes | xs, sm, base, lg | md, xl, 2xl |
| Typography – weights | (none) | regular, medium, semibold, bold |
| Typography – line-heights | (none) | tight, normal, relaxed |
| Spacing | xs, sm, md, lg, xl | Numeric scale (space-1 through space-16) |
| Border radius | sm, md, lg | xs, xl, none, full |
| Shadows | (none) | sm, md, lg, xl |
| Motion | (none) | duration-*, easing-* |
| Z-index | (none) | full scale |
| Breakpoints | (none — in Tailwind only) | CSS env/media tokens |
| Component tokens | (none explicit) | Input heights, button heights, focus ring |
