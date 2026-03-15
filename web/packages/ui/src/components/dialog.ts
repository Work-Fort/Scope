import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export interface DialogOptions {
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'default' | 'danger';
}

export interface DialogResult {
  confirmed: boolean;
}

/**
 * `<wf-dialog>` — Alert and confirm dialog with a promise-based API.
 *
 * Usage:
 *   const dialog = document.querySelector('wf-dialog');
 *   const result = await dialog.show({ message: 'Are you sure?', confirmLabel: 'Yes' });
 *   if (result.confirmed) { ... }
 *
 * For alert-only (no cancel button):
 *   await dialog.alert({ message: 'Done!', confirmLabel: 'OK' });
 *
 * @element wf-dialog
 */
export class WfDialog extends WfElement {
  @property({ type: Boolean, reflect: true }) open = false;

  private _container: HTMLElement | null = null;
  private _resolve: ((result: DialogResult) => void) | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-dialog');
    this._ensureContainer();
  }

  private _ensureContainer(): void {
    if (!this._container) {
      this._container = document.createElement('div');
      this._container.style.display = 'contents';
      this.appendChild(this._container);
    }
  }

  /**
   * Show a confirm dialog. Returns a promise that resolves when the user confirms or cancels.
   */
  show(options: DialogOptions): Promise<DialogResult> {
    return new Promise((resolve) => {
      this._resolve = resolve;
      this._buildDialog(options, true);
      this.open = true;
    });
  }

  /**
   * Show an alert dialog (no cancel button). Returns a promise that resolves when dismissed.
   */
  alert(options: Omit<DialogOptions, 'cancelLabel'>): Promise<DialogResult> {
    return new Promise((resolve) => {
      this._resolve = resolve;
      this._buildDialog({ ...options, cancelLabel: undefined }, false);
      this.open = true;
    });
  }

  /** Close the dialog programmatically. */
  close(confirmed = false): void {
    this.open = false;
    if (this._container) this._container.innerHTML = '';
    if (this._resolve) {
      this._resolve({ confirmed });
      this._resolve = null;
    }
  }

  private _buildDialog(options: DialogOptions, showCancel: boolean): void {
    this._ensureContainer();
    if (!this._container) return;
    this._container.innerHTML = '';

    const variant = options.variant ?? 'default';

    // Overlay
    const overlay = document.createElement('div');
    overlay.className = 'wf-dialog-overlay';
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) this.close(false);
    });

    // Panel
    const panel = document.createElement('div');
    panel.className = 'wf-dialog__panel';
    panel.setAttribute('role', 'alertdialog');
    panel.setAttribute('aria-modal', 'true');
    if (options.title) panel.setAttribute('aria-labelledby', 'wf-dialog-title');
    panel.setAttribute('aria-describedby', 'wf-dialog-message');

    // Title
    if (options.title) {
      const title = document.createElement('h2');
      title.id = 'wf-dialog-title';
      title.className = 'wf-dialog__title';
      title.textContent = options.title;
      panel.appendChild(title);
    }

    // Message
    const message = document.createElement('p');
    message.id = 'wf-dialog-message';
    message.className = 'wf-dialog__message';
    message.textContent = options.message;
    panel.appendChild(message);

    // Actions
    const actions = document.createElement('div');
    actions.className = 'wf-dialog__actions';

    if (showCancel) {
      const cancelBtn = document.createElement('button');
      cancelBtn.className = 'wf-dialog__btn';
      cancelBtn.textContent = options.cancelLabel ?? 'Cancel';
      cancelBtn.addEventListener('click', () => this.close(false));
      actions.appendChild(cancelBtn);
    }

    const confirmBtn = document.createElement('button');
    const btnClass = variant === 'danger' ? 'wf-dialog__btn wf-dialog__btn--danger' : 'wf-dialog__btn wf-dialog__btn--primary';
    confirmBtn.className = btnClass;
    confirmBtn.textContent = options.confirmLabel ?? 'OK';
    confirmBtn.addEventListener('click', () => this.close(true));
    actions.appendChild(confirmBtn);

    panel.appendChild(actions);
    overlay.appendChild(panel);
    this._container.appendChild(overlay);

    // Focus the confirm button
    requestAnimationFrame(() => confirmBtn.focus());

    // Handle Escape key
    this._handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') this.close(false);
    };
    document.addEventListener('keydown', this._handleEscape);
  }

  private _handleEscape: ((e: KeyboardEvent) => void) | null = null;

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._handleEscape) {
      document.removeEventListener('keydown', this._handleEscape);
      this._handleEscape = null;
    }
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('open') && !this.open && this._handleEscape) {
      document.removeEventListener('keydown', this._handleEscape);
      this._handleEscape = null;
    }
  }
}

customElements.define('wf-dialog', WfDialog);

declare global {
  interface HTMLElementTagNameMap {
    'wf-dialog': WfDialog;
  }
}
