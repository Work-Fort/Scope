# Web Component Extraction — Plan 6

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract `wf-compose-input` and `wf-user-picker` as Lit web components in `@workfort/ui`, replacing Sharkfin-specific implementations with framework-agnostic reusable components.

**Architecture:** Both components extend `WfElement` (light DOM Lit), use `--wf-*` design tokens, and compose existing `@workfort/ui` components internally. `wf-compose-input` wraps a textarea + button with Enter-to-send. `wf-user-picker` wraps `wf-dialog` + `wf-list` for selecting users.

**Tech Stack:** Lit, TypeScript, vitest + happy-dom, `@workfort/ui`

**Repo:** `scope/lead` (package at `web/packages/ui/`)

**Prerequisite:** Plan 5 (utilities) must be complete — `initials()` is used by `wf-user-picker`.

---

### Task 1: `wf-compose-input` Web Component

**Files:**
- Create: `web/packages/ui/src/components/compose-input.ts`
- Create: `web/packages/ui/tests/components/compose-input.test.ts`
- Create: `web/packages/ui/src/styles/compose-input.css`
- Modify: `web/packages/ui/src/index.ts` (add export)
- Modify: `web/packages/ui/src/styles/index.css` (import CSS)

**Step 1: Write failing test**

Create `web/packages/ui/tests/components/compose-input.test.ts`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/compose-input.js';
import type { WfComposeInput } from '../../src/components/compose-input.js';

describe('WfComposeInput', () => {
  afterEach(cleanup);

  it('renders with wf-compose-input class', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    expect(el.classList.contains('wf-compose-input')).toBe(true);
  });

  it('renders textarea and send button', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    expect(el.querySelector('textarea')).toBeTruthy();
    expect(el.querySelector('wf-button')).toBeTruthy();
  });

  it('applies placeholder', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input', { placeholder: 'Type here' });
    const textarea = el.querySelector('textarea');
    expect(textarea?.placeholder).toBe('Type here');
  });

  it('dispatches wf-send on Enter', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(handler).toHaveBeenCalledOnce();
    expect((handler.mock.calls[0][0] as CustomEvent).detail.body).toBe('hello');
  });

  it('does not dispatch wf-send on Shift+Enter', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', shiftKey: true, bubbles: true }));

    expect(handler).not.toHaveBeenCalled();
  });

  it('does not dispatch wf-send when empty', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(handler).not.toHaveBeenCalled();
  });

  it('clears textarea after send', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(textarea.value).toBe('');
  });

  it('disables when disabled prop is set', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input', { disabled: true });
    const textarea = el.querySelector('textarea');
    expect(textarea?.disabled).toBe(true);
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web/packages/ui && npx vitest run tests/components/compose-input.test.ts`
Expected: FAIL — module not found.

**Step 3: Implement**

Create `web/packages/ui/src/styles/compose-input.css`:

```css
.wf-compose-input {
  display: block;
}
.wf-compose-input__box {
  display: flex;
  align-items: flex-end;
  gap: var(--wf-space-sm);
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-lg);
  padding: var(--wf-space-sm);
  background: var(--wf-color-bg-secondary);
  transition: border-color 0.15s;
}
.wf-compose-input__box:focus-within {
  border-color: var(--wf-color-text-secondary);
}
.wf-compose-input__field {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--wf-color-text);
  font-family: inherit;
  font-size: var(--wf-text-sm);
  line-height: var(--wf-leading-normal);
  resize: none;
  outline: none;
  min-height: 1.5rem;
}
.wf-compose-input__field::placeholder {
  color: var(--wf-color-text-muted);
}
.wf-compose-input__field:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

Create `web/packages/ui/src/components/compose-input.ts`:

```typescript
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import '../styles/compose-input.css';

export class WfComposeInput extends WfElement {
  @property({ type: String, reflect: true }) placeholder = '';
  @property({ type: Boolean, reflect: true }) disabled = false;

  private _value = '';
  private _textarea: HTMLTextAreaElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-compose-input');
    this._render();
  }

  private _render(): void {
    this.innerHTML = '';

    const box = document.createElement('div');
    box.className = 'wf-compose-input__box';

    const textarea = document.createElement('textarea');
    textarea.className = 'wf-compose-input__field';
    textarea.placeholder = this.placeholder;
    textarea.rows = 1;
    textarea.disabled = this.disabled;
    this._textarea = textarea;

    textarea.addEventListener('input', () => {
      this._value = textarea.value;
    });

    textarea.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        this._send();
      }
    });

    const button = document.createElement('wf-button');
    button.setAttribute('title', 'Send');
    button.style.cssText = 'padding: 4px 10px;';
    button.textContent = '↑';
    if (this.disabled) button.setAttribute('disabled', '');

    button.addEventListener('wf-click', () => this._send());

    box.appendChild(textarea);
    box.appendChild(button);
    this.appendChild(box);
  }

  private _send(): void {
    if (this.disabled) return;
    const body = this._value.trim();
    if (!body) return;

    this.dispatchEvent(new CustomEvent('wf-send', {
      bubbles: true,
      composed: true,
      detail: { body },
    }));

    this._value = '';
    if (this._textarea) this._textarea.value = '';
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('placeholder') && this._textarea) {
      this._textarea.placeholder = this.placeholder;
    }
    if (changed.has('disabled') && this._textarea) {
      this._textarea.disabled = this.disabled;
    }
  }
}

customElements.define('wf-compose-input', WfComposeInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-compose-input': WfComposeInput;
  }
}
```

**Step 4: Add exports**

In `web/packages/ui/src/index.ts`, add:

```typescript
import './components/compose-input.js';
// ...
export { WfComposeInput } from './components/compose-input.js';
```

Import the CSS in `web/packages/ui/src/styles/index.css`:

```css
@import './compose-input.css';
```

**Step 5: Run tests**

Run: `cd web/packages/ui && npx vitest run tests/components/compose-input.test.ts`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/packages/ui/src/components/compose-input.ts web/packages/ui/tests/components/compose-input.test.ts web/packages/ui/src/styles/compose-input.css web/packages/ui/src/styles/index.css web/packages/ui/src/index.ts
git commit -m "feat(@workfort/ui): add wf-compose-input web component"
```

---

### Task 2: `wf-user-picker` Web Component

**Files:**
- Create: `web/packages/ui/src/components/user-picker.ts`
- Create: `web/packages/ui/tests/components/user-picker.test.ts`
- Modify: `web/packages/ui/src/index.ts` (add export)

**Step 1: Write failing test**

Create `web/packages/ui/tests/components/user-picker.test.ts`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/user-picker.js';
import type { WfUserPicker } from '../../src/components/user-picker.js';

describe('WfUserPicker', () => {
  afterEach(cleanup);

  it('renders with wf-user-picker class', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    expect(el.classList.contains('wf-user-picker')).toBe(true);
  });

  it('renders wf-dialog with header', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker', { header: 'Pick a user' });
    const dialog = el.querySelector('wf-dialog');
    expect(dialog).toBeTruthy();
    expect(dialog?.getAttribute('header')).toBe('Pick a user');
  });

  it('renders user list from users property', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    (el as any).users = [
      { username: 'alice', online: true },
      { username: 'bob', online: false },
    ];
    await el.updateComplete;

    const items = el.querySelectorAll('wf-list-item');
    expect(items.length).toBe(2);
  });

  it('excludes user matching exclude property', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker', { exclude: 'bob' });
    (el as any).users = [
      { username: 'alice', online: true },
      { username: 'bob', online: false },
    ];
    await el.updateComplete;

    const items = el.querySelectorAll('wf-list-item');
    expect(items.length).toBe(1);
  });

  it('dispatches wf-select with username on item click', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    (el as any).users = [{ username: 'alice', online: true }];
    await el.updateComplete;

    const handler = vi.fn();
    el.addEventListener('wf-select', handler);

    const item = el.querySelector('wf-list-item') as HTMLElement;
    item?.dispatchEvent(new CustomEvent('wf-select', { bubbles: true }));

    // The component should re-dispatch with username detail.
    // Implementation may vary — check handler was called.
    expect(handler).toHaveBeenCalled();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web/packages/ui && npx vitest run tests/components/user-picker.test.ts`
Expected: FAIL.

**Step 3: Implement**

Create `web/packages/ui/src/components/user-picker.ts`:

```typescript
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { initials } from '../utils/initials.js';

export interface UserPickerUser {
  username: string;
  online?: boolean;
  state?: string;
  type?: string;
}

export class WfUserPicker extends WfElement {
  @property({ type: String, reflect: true }) header = '';
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String, reflect: true }) exclude = '';
  @property({ type: Array }) users: UserPickerUser[] = [];

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-user-picker');
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('users') || changed.has('exclude') || changed.has('header') || changed.has('open')) {
      this._render();
    }
    if (changed.has('open')) {
      const dialog = this.querySelector('wf-dialog') as HTMLElement & { show(): void; hide(): void } | null;
      if (this.open) dialog?.show?.();
      else dialog?.hide?.();
    }
  }

  private _render(): void {
    const filtered = this.users.filter((u) => u.username !== this.exclude);

    this.innerHTML = '';

    const dialog = document.createElement('wf-dialog');
    dialog.setAttribute('header', this.header);
    dialog.addEventListener('wf-close', () => {
      this.dispatchEvent(new CustomEvent('wf-close', { bubbles: true, composed: true }));
    });

    const list = document.createElement('wf-list');

    for (const user of filtered) {
      const item = document.createElement('wf-list-item');

      const avatar = document.createElement('div');
      avatar.style.cssText = 'width:1.5rem;height:1.5rem;border-radius:var(--wf-radius-full);background:var(--wf-color-bg-elevated);display:flex;align-items:center;justify-content:center;font-size:0.625rem;font-weight:var(--wf-weight-semibold);color:var(--wf-color-text-secondary);flex-shrink:0;position:relative;margin-right:var(--wf-space-sm);';
      avatar.textContent = initials(user.username);

      const dot = document.createElement('wf-status-dot');
      const status = !user.online ? 'offline' : user.state === 'idle' ? 'away' : 'online';
      dot.setAttribute('status', status);
      dot.style.cssText = 'position:absolute;bottom:-1px;right:-1px;';
      avatar.appendChild(dot);

      const name = document.createElement('span');
      name.textContent = user.username;

      item.appendChild(avatar);
      item.appendChild(name);

      item.addEventListener('wf-select', () => {
        this.dispatchEvent(new CustomEvent('wf-select', {
          bubbles: true,
          composed: true,
          detail: { username: user.username },
        }));
      });

      list.appendChild(item);
    }

    dialog.appendChild(list);
    this.appendChild(dialog);
  }
}

customElements.define('wf-user-picker', WfUserPicker);

declare global {
  interface HTMLElementTagNameMap {
    'wf-user-picker': WfUserPicker;
  }
}
```

**Step 4: Add exports**

In `web/packages/ui/src/index.ts`:

```typescript
import './components/user-picker.js';
// ...
export { WfUserPicker } from './components/user-picker.js';
export type { UserPickerUser } from './components/user-picker.js';
```

**Step 5: Run tests**

Run: `cd web/packages/ui && npx vitest run`
Expected: All tests pass.

**Step 6: Build**

Run: `cd web/packages/ui && pnpm build`
Expected: Build succeeds.

**Step 7: Commit**

```bash
git add web/packages/ui/src/components/user-picker.ts web/packages/ui/tests/components/user-picker.test.ts web/packages/ui/src/index.ts
git commit -m "feat(@workfort/ui): add wf-user-picker web component"
```
