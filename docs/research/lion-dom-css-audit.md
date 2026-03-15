# Lion DOM vs CSS Selector Audit

**Date:** 2026-03-14
**File audited:** `web/packages/ui/src/styles/forms.css`
**Method:** Playwright inspection of live Storybook stories at `localhost:6006`

---

## Summary

The CSS file assumes a DOM structure that partially diverges from what Lion actually renders. The key finding is that **`.wf-field__container` does not exist** in any Lion component's DOM output. Lion uses its own internal wrapper structure (`input-group__container`, `form-field__group-one`, etc.) which our CSS never targets. Additionally, checkbox and radio inputs receive unintended styles from the generic `.wf-field__input` rules.

---

## 1. wf-input

### Actual DOM (Lion renders)
```
<wf-input class="form-field wf-field wf-input">
  <label slot="label" class="wf-field__label">...</label>
  <div slot="help-text" class="wf-field__help-text">...</div>
  <lion-validation-feedback slot="feedback" class="wf-field__feedback">...</lion-validation-feedback>
  <input slot="input" class="form-control wf-field__input" type="text">
  <!-- Shadow-DOM-like light DOM template: -->
  <div class="form-field__group-one">
    <div class="form-field__label"><slot name="label"></div>
    <small class="form-field__help-text"><slot name="help-text"></small>
  </div>
  <div class="form-field__group-two">
    <div class="input-group">
      <div class="input-group__before"><slot name="before"></div>
      <div class="input-group__container">
        <div class="input-group__input"><slot name="input"></div>
      </div>
      <div class="input-group__after"><slot name="after"></div>
    </div>
    <div class="form-field__feedback"><slot name="feedback"></div>
  </div>
</wf-input>
```

### CSS targeting analysis

| CSS Selector | Matches? | Notes |
|---|---|---|
| `.wf-field` | YES | Applied via class on `<wf-input>` |
| `.wf-field__label` | YES | Applied via class on `<label slot="label">` |
| `.wf-field__label--required::after` | CONDITIONAL | Only if component adds this class (not seen in default story) |
| **`.wf-field__container`** | **NO** | Lion renders `input-group__container` instead. This is dead CSS for input. |
| **`.wf-field__container:focus-within`** | **NO** | Dead CSS. |
| `.wf-field--error .wf-field__container` | **NO** | Dead CSS -- container doesn't exist. |
| `.wf-field--disabled` | CONDITIONAL | Only if component adds this class |
| `.wf-field__input` | YES | Applied via class on `<input slot="input">` |
| **`.wf-field__container .wf-field__input`** | **NO** | Dead CSS -- no `.wf-field__container` ancestor exists |
| `.wf-field__input:focus` | YES | Works on the `<input>` element |
| `.wf-field--error .wf-field__input` | CONDITIONAL | Only if error class is added |
| `.wf-field__input::placeholder` | YES | Works |
| `.wf-field__help-text` | YES | Applied via class on `<div slot="help-text">` |
| `.wf-field__feedback` | YES | Applied via class on `<lion-validation-feedback>` |

### Issues found
1. **Dead CSS:** `.wf-field__container` and all rules depending on it never match. The input gets its own border from `.wf-field__input` (lines 57-70), which works. But the container-based border delegation pattern (lines 33-46, 72-78) is entirely dead.
2. **Lion internal elements unstyled (benign):** `form-field__group-one`, `form-field__group-two`, `input-group`, `input-group__container`, `input-group__before`, `input-group__after`, `input-group__input` -- these are Lion's internal layout wrappers. They likely get default styles from Lion's base CSS, but we have no overrides for them. This appears to be working correctly since our `wf-field__input` styles are sufficient.

---

## 2. wf-textarea

### Actual DOM (Lion renders)
Same structure as `wf-input`, except:
- Uses `<textarea slot="input" class="form-control wf-field__input">` instead of `<input>`
- Host element has classes `form-field wf-field wf-textarea`
- Same Lion internal wrapper structure (`form-field__group-one/two`, `input-group__container`, etc.)

### CSS targeting analysis

| CSS Selector | Matches? | Notes |
|---|---|---|
| **`.wf-textarea .wf-field__container`** | **NO** | Dead CSS (line 107-110). No `.wf-field__container` exists. |
| `.wf-textarea .wf-field__input` | YES | Matches `<textarea class="wf-field__input">` |

### Issues found
1. **Dead CSS:** `.wf-textarea .wf-field__container` (line 107) sets `height: auto` and padding, but the container doesn't exist. The textarea gets sized by its own `rows` attribute and Lion's inline styles (`height`, `max-height`, `overflow`).
2. **Textarea resize:** CSS sets `resize: vertical` (line 113) on `.wf-textarea .wf-field__input`, but Lion sets `resize: none` via inline style on the textarea element. **The inline style wins**, so `resize: vertical` in our CSS is effectively dead.

---

## 3. wf-select

### Actual DOM (Lion renders)
Same structure as `wf-input`, except:
- Uses `<select slot="input" class="form-control wf-field__input">` with `<option>` children
- Host element has classes `form-field wf-field wf-select`

### CSS targeting analysis

| CSS Selector | Matches? | Notes |
|---|---|---|
| `.wf-select .wf-field__input` | YES | Matches the `<select>` element |

### Issues found
1. **No critical issues.** The select-specific CSS (lines 118-127) all targets `.wf-select .wf-field__input`, which correctly matches. The `appearance: none`, custom arrow background-image, and padding-right all apply.
2. **Minor:** The select inherits the generic `.wf-field__input` `width: 100%` and border styles, which is correct behavior.

---

## 4. wf-checkbox

### Actual DOM (Lion renders)
```
<wf-checkbox class="form-field wf-field wf-checkbox">
  <label slot="label" class="wf-field__label">Accept terms</label>
  <div slot="help-text" class="wf-field__help-text">...</div>
  <lion-validation-feedback slot="feedback" class="wf-field__feedback">...</lion-validation-feedback>
  <input slot="input" class="form-control wf-field__input" type="checkbox">
  <!-- Choice field template (different from input/textarea/select!): -->
  <slot name="input"></slot>
  <div class="choice-field__graphic-container" aria-hidden="true">...</div>
  <div class="choice-field__label"><slot name="label"></div>
  <small class="choice-field__help-text"><slot name="help-text"></small>
</wf-checkbox>
```

**Key difference:** Checkbox uses a completely different Lion template -- no `form-field__group-one/two`, no `input-group` wrappers. Instead it has `choice-field__graphic-container`, `choice-field__label`, and `choice-field__help-text`.

### CSS targeting analysis

| CSS Selector | Matches? | Notes |
|---|---|---|
| `.wf-checkbox` | YES | Host element |
| `.wf-checkbox .wf-field__input[type="checkbox"]` | YES | Matches the hidden `<input type="checkbox">` |
| `.wf-checkbox .choice-field__graphic-container` | YES | Matches Lion's graphic container |
| `.wf-checkbox > [slot="label"]` | YES | Matches `<label slot="label">` |
| **`.wf-checkbox .wf-field__input:checked ~ .choice-field__graphic-container`** | **MAYBE** | See issue below |
| `.wf-checkbox .choice-field__label` | YES | Matches, hidden via `display: none` |
| `.wf-checkbox .choice-field__help-text` | YES | Matches, hidden via `display: none` |
| `.wf-checkbox .form-field__group-one` | **NO** | Dead CSS -- checkbox template doesn't render these |
| `.wf-checkbox .form-field__group-two` | **NO** | Dead CSS -- checkbox template doesn't render these |

### Issues found

1. **CRITICAL -- Unintended styles on checkbox input:** The generic `.wf-field__input` rule (lines 57-70) applies to the checkbox `<input>` because it has class `wf-field__input`. This means the checkbox input gets:
   - `flex: 1` -- meaningless since it's `position: absolute`
   - `width: 100%` -- overridden by `width: 1px` in the checkbox-specific rule
   - `border: 1px solid var(--wf-color-border-strong)` -- overridden by `border: none`
   - `border-radius`, `padding`, `background`, `font-size` -- all overridden

   The checkbox-specific rule (lines 141-149) **does override** these, but this creates unnecessary specificity conflicts and fragility. If the order of CSS rules changes, the checkbox could break.

2. **Sibling selector concern:** `.wf-checkbox .wf-field__input:checked ~ .choice-field__graphic-container` (line 171) uses the general sibling combinator `~`. In the DOM, the `<input>` (slotted) and the `<div class="choice-field__graphic-container">` (template) are both direct children of `<wf-checkbox>`. However, between them is a `<slot name="input">` element. The `~` combinator should still work because it selects any following sibling, not just adjacent ones. **This selector appears correct** in light DOM rendering.

3. **Dead CSS for hiding Lion internals:** `.wf-checkbox .form-field__group-one` and `.wf-checkbox .form-field__group-two` (line 181-182) target elements that don't exist in the checkbox template. These rules are dead CSS (though harmless).

---

## 5. wf-radio

### Actual DOM (Lion renders)
```
<wf-radio-group class="form-control wf-radio-group" role="radiogroup">
  <wf-radio class="form-field wf-field wf-radio">
    <label slot="label" class="wf-field__label">Free</label>
    <div slot="help-text" class="wf-field__help-text">...</div>
    <lion-validation-feedback slot="feedback" class="wf-field__feedback">...</lion-validation-feedback>
    <input value="free" slot="input" class="form-control wf-field__input" type="radio">
    <slot name="input"></slot>
    <div class="choice-field__graphic-container" aria-hidden="true">...</div>
    <div class="choice-field__label"><slot name="label"></div>
    <small class="choice-field__help-text"><slot name="help-text"></small>
  </wf-radio>
  <!-- more wf-radio children... -->
</wf-radio-group>
```

Identical internal structure to checkbox (choice field template).

### CSS targeting analysis

| CSS Selector | Matches? | Notes |
|---|---|---|
| `.wf-radio` | YES | Host element |
| `.wf-radio .wf-field__input[type="radio"]` | YES | Matches the hidden `<input type="radio">` |
| `.wf-radio .choice-field__graphic-container` | YES | Matches |
| `.wf-radio > [slot="label"]` | YES | Matches |
| `.wf-radio .wf-field__input:checked ~ .choice-field__graphic-container` | YES | Same structure as checkbox, sibling combinator works |
| `.wf-radio .choice-field__label` | YES | Matches, hidden via `display: none` |
| `.wf-radio .choice-field__help-text` | YES | Matches, hidden via `display: none` |
| `.wf-radio .form-field__group-one` | **NO** | Dead CSS |
| `.wf-radio .form-field__group-two` | **NO** | Dead CSS |
| `.wf-radio-group` | YES | Matches the group wrapper |

### Issues found

1. **Same unintended-styles issue as checkbox:** The generic `.wf-field__input` rules (lines 57-70) apply to the radio `<input>` -- `width: 100%`, `border`, `flex: 1`, etc. The radio-specific visually-hidden rule (lines 205-213) overrides these, but it's fragile.

2. **Dead CSS:** `.wf-radio .form-field__group-one` and `.wf-radio .form-field__group-two` (lines 245-246) don't match anything.

---

## Consolidated Findings

### Dead CSS Rules (target non-existent elements)

| Line(s) | Selector | Reason |
|---|---|---|
| 33-42 | `.wf-field__container` | Lion never renders this element. Uses `input-group__container` instead. |
| 44-46 | `.wf-field__container:focus-within` | Container doesn't exist. |
| 48-50 | `.wf-field--error .wf-field__container` | Container doesn't exist. |
| 73-78 | `.wf-field__container .wf-field__input` | No container ancestor exists. |
| 107-110 | `.wf-textarea .wf-field__container` | Container doesn't exist in textarea either. |
| 113 | `.wf-textarea .wf-field__input` `resize: vertical` | Overridden by Lion's inline `resize: none`. |
| 181 | `.wf-checkbox .form-field__group-one` | Checkbox template doesn't include these. |
| 182 | `.wf-checkbox .form-field__group-two` | Checkbox template doesn't include these. |
| 245 | `.wf-radio .form-field__group-one` | Radio template doesn't include these. |
| 246 | `.wf-radio .form-field__group-two` | Radio template doesn't include these. |

### Unintended Styles (CSS applies to wrong elements)

| Line(s) | Selector | Problem |
|---|---|---|
| 57-70 | `.wf-field__input` | Applies `width: 100%`, `flex: 1`, `border`, `padding`, `background` to checkbox and radio `<input>` elements. These are overridden by component-specific rules, but the generic rule applies first and creates fragile specificity dependencies. |

### Lion-rendered Elements With No Custom CSS (potential gaps)

| Element | Components | Notes |
|---|---|---|
| `.form-field__group-one` | input, textarea, select | Lion's internal layout wrapper. Probably styled by Lion's base CSS. Not a problem unless we need to override. |
| `.form-field__group-two` | input, textarea, select | Same as above. |
| `.input-group` | input, textarea, select | Lion's input group wrapper. |
| `.input-group__container` | input, textarea, select | This is what Lion uses instead of our expected `.wf-field__container`. |
| `.input-group__before` | input, textarea, select | Slot wrapper for prefix content. |
| `.input-group__after` | input, textarea, select | Slot wrapper for suffix content. |
| `.input-group__input` | input, textarea, select | Immediate wrapper around the slotted input. |
| `.form-field__label` | input, textarea, select | Lion's internal label wrapper (distinct from our slotted `<label>`). |
| `.form-field__help-text` | input, textarea, select | Lion's internal help-text wrapper. |
| `.form-field__feedback` | input, textarea, select | Lion's internal feedback wrapper. |

### Recommendations

1. **Remove `.wf-field__container` rules entirely** or replace them with selectors targeting Lion's actual `input-group__container` if container-level border styling is desired.
2. **Scope `.wf-field__input` generic rules** to exclude checkbox/radio, e.g., `.wf-field__input:not([type="checkbox"]):not([type="radio"])`, to prevent unintended style leakage.
3. **Remove dead `form-field__group-one/two` selectors** from checkbox and radio sections (lines 181-182, 245-246).
4. **Investigate textarea resize:** Either remove `resize: vertical` from CSS (since Lion overrides it) or find a way to override Lion's inline style if vertical resize is desired.
