# Visual Design Review -- WorkFort UI Storybook

**Date:** 2026-03-15
**Reviewer:** Automated (Claude + Playwright)
**Storybook URL:** http://localhost:6006
**Stories reviewed:** 110 stories across 4 categories (Components, Forms, Layout, Navigation)
**Modes tested:** Dark (default) and Light (data-theme="light")

---

## 1. Summary

**Overall consistency rating: 4/10**

The component library has a solid foundation but suffers from three systemic issues that affect nearly every component:

1. **Slot ordering bug** -- Slotted content (body/children) renders ABOVE headers/titles in multiple components (Card, Panel, Accordion, Dialog, Drawer). This is a structural problem in the web component slot rendering.
2. **Dark mode contrast** -- Borders, dividers, tracks, and subtle backgrounds are nearly invisible in dark mode. The `--wf-color-border` token does not provide sufficient contrast against `--wf-color-bg` in the dark theme.
3. **Tabs component is completely broken** -- Renders nothing at all (no tab list, all panels hidden).

Font consistency is good throughout (sans-serif). Light mode is generally more readable. Color semantics (error=red, warning=amber, success=green) are consistently applied.

---

## 2. Issues Found

### Critical

| Component | Mode | Issue | Severity |
|-----------|------|-------|----------|
| Tabs | Both | Component renders completely blank. No tab buttons rendered, all tab panels have `display: none`. The tab list is never generated. | Critical |
| Dialog | Both | Body content renders OUTSIDE the dialog panel (appears behind it in the page flow). Dialog only shows title + close button, no body inside. | Critical |
| Drawer | Both | Body content renders OUTSIDE the drawer panel (appears in page background). Drawer panel only shows title + close button. | Critical |

### Major

| Component | Mode | Issue | Severity |
|-----------|------|-------|----------|
| Card (all variants) | Both | Body content renders ABOVE the card title/header. Expected: title first, then body. | Major |
| Panel | Both | Body content renders ABOVE the panel title. Same slot ordering issue as Card. | Major |
| Accordion | Both | Answer content renders ABOVE each question header button. Content should appear below the trigger. | Major |
| Popover | Both | Popover content has no visual container -- no background, border, or shadow. Content just floats in the page like inline text. | Major |
| Toast (info) | Dark | Toast background is nearly identical to page background. No visible boundary or container. Almost invisible. | Major |
| Toast (success) | Dark | Dark text on dark green background -- poor text contrast ratio. | Major |
| Radio group | Both | Group label ("Plan") renders BELOW the radio options instead of above them. | Major |
| Checkbox group | Both | Group label ("Interests") renders BELOW the checkbox options instead of above them. | Major |

### Minor

| Component | Mode | Issue | Severity |
|-----------|------|-------|----------|
| Input, Select, Combobox, DatePicker, Textarea | Dark | Input field borders are barely visible (very low contrast against dark background). | Minor |
| Divider | Dark | Separator line is nearly invisible. | Minor |
| Skeleton | Dark | Skeleton placeholder bar has very low contrast against background. | Minor |
| Spinner | Dark | Spinner circles are barely visible -- very low contrast against dark background. | Minor |
| Slider | Dark | Track line is barely visible. No visible label text (only in accessibility tree). | Minor |
| Toggle (unchecked) | Dark | Toggle track is barely visible when unchecked. | Minor |
| Toggle (checked) | Dark | Toggle track is still very dark when checked -- hard to distinguish on/off state. | Minor |
| Card (outlined) | Dark | Card outline/border barely visible against dark background. | Minor |
| ScrollArea | Dark | Container border barely visible. | Minor |
| Banner (info) | Both | Banner background barely differs from page background. Left accent border is subtle. | Minor |
| Toast container | Dark | Toast notification has minimal visual distinction from background. | Minor |
| Breadcrumbs | Dark | Separator characters between crumbs are very subtle. | Minor |

---

## 3. Recommendations

### P0 -- Fix immediately

1. **Fix Tabs component** -- The component is non-functional. Tab buttons are not being rendered and all panels are hidden. Investigate the `connectedCallback` / `firstUpdated` lifecycle to ensure the tab list is generated from child `wf-tab-panel` elements.

2. **Fix slot ordering in Card, Panel, Accordion, Dialog, Drawer** -- Slotted content consistently renders in the wrong order. The header/title slot should render before the body/default slot. Check the `render()` method template ordering in these components -- the named slot (header/title) should come first in the template, followed by the default slot (body/content).

3. **Fix Dialog/Drawer content slotting** -- Body content is not being projected into the dialog/drawer panel. The default slot content appears outside the overlay. The story may be placing content outside the component's slot boundary, or the component's slot projection is broken.

### P1 -- Fix soon

4. **Fix Popover container styling** -- Add a visible background, border, border-radius, padding, and box-shadow to the popover content panel. Currently the popover content has no visual enclosure.

5. **Fix Radio/Checkbox group label position** -- The group label should render ABOVE the options, not below. Check the fieldset/legend or slot ordering in `wf-radio-group` and `wf-checkbox-group`.

6. **Increase dark mode border contrast** -- The `--wf-color-border` token in dark mode needs more contrast. Current value is too close to `--wf-color-bg`. Recommend increasing lightness from ~15% to ~25-30% to make borders clearly visible on dark backgrounds. This single token change would fix borders on: Input, Select, Combobox, DatePicker, Textarea, Card, Panel, Divider, ScrollArea, and Table.

### P2 -- Improve when possible

7. **Improve Toast visibility in dark mode** -- The info toast needs a more distinct background or border. Consider using `--wf-color-surface` or adding a subtle border to all toast variants.

8. **Improve Skeleton contrast in dark mode** -- Increase the skeleton shimmer/placeholder brightness to be more visible.

9. **Improve Spinner contrast in dark mode** -- Use a brighter color for spinner circles or add a contrasting stroke.

10. **Improve Toggle track visibility** -- The toggle track is too dark in both checked and unchecked states in dark mode. The checked state should use `--wf-color-primary` or a clearly visible accent color.

11. **Improve Slider track and label** -- Make the slider track more visible in dark mode and render the label text visually (currently only in the accessibility tree).

---

## 4. Components That Look Good

The following components rendered correctly in both themes with no significant issues:

- **Badge** -- Renders correctly, good contrast in both modes
- **Button (text, filled, disabled)** -- Good contrast, proper styling
- **AlertDialog** -- Properly centered, good contrast, buttons readable
- **Stepper** -- Good color coding (green=done, outlined=current, gray=future)
- **Progress** -- Clear bar, good contrast, readable label and percentage
- **Pagination** -- Active page clearly highlighted, buttons readable
- **Breadcrumbs** -- Links readable, proper hierarchy
- **StatusDot** -- Colors (green/gray/amber) clearly visible
- **Tooltip** -- Properly positioned, readable text, good contrast
- **Table** -- Readable columns, proper alignment, pagination works
- **List** -- Active item highlighted, readable text
- **ErrorFallback** -- Bold title, readable subtitle
- **FileUpload** -- Dashed border visible, centered text readable
- **Checkbox** -- Properly styled, label aligned with checkbox
- **Banner (warning, error)** -- Good color coding with accent borders
