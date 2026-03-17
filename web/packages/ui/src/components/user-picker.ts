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

  private _container: HTMLDivElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-user-picker');
    this._container = document.createElement('div');
    this.appendChild(this._container);
    this._renderContent();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('users') || changed.has('exclude') || changed.has('header') || changed.has('open')) {
      this._renderContent();
    }
    if (changed.has('open')) {
      const dialog = this._container?.querySelector('wf-dialog') as HTMLElement & { show(): void; hide(): void } | null;
      if (this.open) dialog?.show?.();
      else dialog?.hide?.();
    }
  }

  private _renderContent(): void {
    if (!this._container) return;

    const filtered = this.users.filter((u) => u.username !== this.exclude);

    this._container.innerHTML = '';

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
    this._container.appendChild(dialog);
  }
}

customElements.define('wf-user-picker', WfUserPicker);

declare global {
  interface HTMLElementTagNameMap {
    'wf-user-picker': WfUserPicker;
  }
}
