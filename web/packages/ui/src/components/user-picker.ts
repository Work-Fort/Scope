import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

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

      const avatar = document.createElement('wf-avatar');
      avatar.setAttribute('username', user.username);
      avatar.setAttribute('size', 'sm');
      if (user.online) {
        avatar.setAttribute('status', user.state === 'idle' ? 'away' : 'online');
      } else {
        avatar.setAttribute('status', 'offline');
      }
      avatar.style.marginRight = 'var(--wf-space-sm)';

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
