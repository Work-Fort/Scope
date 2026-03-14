// src/form/wf-file-upload.ts
// File upload component with drop zone and native file input.
// Does not use Lion — purely custom with LitElement base.
import { LitElement, html } from 'lit';

/**
 * `<wf-file-upload>` — File upload component with a clickable drop zone.
 * Fires `wf-change` when files are selected.
 *
 * @element wf-file-upload
 * @fires wf-change — When files are selected via click or drop.
 */
export class WfFileUpload extends LitElement {
  static get properties() {
    return {
      accept: { type: String, reflect: true },
      multiple: { type: Boolean, reflect: true },
      disabled: { type: Boolean, reflect: true },
    };
  }

  accept = '';
  multiple = false;
  disabled = false;

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-file-upload');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-file-upload--disabled', this.disabled);
  }

  private _handleClick(): void {
    if (this.disabled) return;
    const input = this.querySelector('.wf-file-upload__input') as HTMLInputElement;
    input?.click();
  }

  private _handleChange(e: Event): void {
    e.stopPropagation();
    const input = e.target as HTMLInputElement;
    const files = input.files;
    this.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { files },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleDragOver(e: DragEvent): void {
    if (this.disabled) return;
    e.preventDefault();
    e.stopPropagation();
  }

  private _handleDrop(e: DragEvent): void {
    if (this.disabled) return;
    e.preventDefault();
    e.stopPropagation();
    const files = e.dataTransfer?.files ?? null;
    this.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { files },
        bubbles: true,
        composed: true,
      }),
    );
  }

  render() {
    return html`
      <input
        type="file"
        class="wf-file-upload__input"
        .accept=${this.accept}
        ?multiple=${this.multiple}
        ?disabled=${this.disabled}
        @change=${this._handleChange}
      />
      <div
        class="wf-file-upload__drop-zone"
        @click=${this._handleClick}
        @dragover=${this._handleDragOver}
        @drop=${this._handleDrop}
      >
        <span class="wf-file-upload__text">
          <slot>Click or drag files to upload</slot>
        </span>
      </div>
    `;
  }
}

customElements.define('wf-file-upload', WfFileUpload);

declare global {
  interface HTMLElementTagNameMap {
    'wf-file-upload': WfFileUpload;
  }
}
