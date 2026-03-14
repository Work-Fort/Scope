// src/form/wf-combobox.ts
// Combobox with text input and filterable dropdown list.
// Custom implementation using LitElement base (no Lion dependency).
import { LitElement, html } from 'lit';

/**
 * `<wf-combobox>` — Combines a text input with a filterable dropdown.
 * Options can be provided via the `options` property (array of
 * `{ value: string; label: string }`) or via `<wf-option>` children.
 *
 * @element wf-combobox
 * @fires wf-change — When the user selects an option.
 * @fires wf-input — When the user types in the input.
 */
export class WfCombobox extends LitElement {
  static get properties() {
    return {
      label: { type: String, reflect: true },
      value: { type: String, reflect: true },
      placeholder: { type: String, reflect: true },
      disabled: { type: Boolean, reflect: true },
      options: { type: Array },
      _open: { type: Boolean },
      _filter: { type: String },
      _highlightedIndex: { type: Number },
    };
  }

  label = '';
  value = '';
  placeholder = '';
  disabled = false;
  options: Array<{ value: string; label: string }> = [];

  _open = false;
  _filter = '';
  _highlightedIndex = -1;

  private _boundDocClick: ((e: Event) => void) | null = null;

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-combobox');
    this._syncClasses();
    this._boundDocClick = this._handleDocClick.bind(this);
    document.addEventListener('click', this._boundDocClick);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._boundDocClick) {
      document.removeEventListener('click', this._boundDocClick);
    }
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
    // Imperatively sync disabled to native input for happy-dom compat
    const input = this.querySelector(
      '.wf-combobox__input',
    ) as HTMLInputElement | null;
    if (input) {
      input.disabled = this.disabled;
    }
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-combobox--disabled', this.disabled);
    this.classList.toggle('wf-combobox--open', this._open);
  }

  get filteredOptions(): Array<{ value: string; label: string }> {
    if (!this._filter) return this.options;
    const lower = this._filter.toLowerCase();
    return this.options.filter((o) => o.label.toLowerCase().includes(lower));
  }

  private _handleDocClick(e: Event): void {
    if (!this.contains(e.target as Node)) {
      this._open = false;
    }
  }

  /** Open the dropdown. Called on input focus. */
  open(): void {
    if (!this.disabled) {
      this._open = true;
    }
  }

  /** Close the dropdown. */
  close(): void {
    this._open = false;
    this._highlightedIndex = -1;
  }

  private _handleInputFocus(): void {
    this.open();
  }

  private _handleInput(e: Event): void {
    const input = e.target as HTMLInputElement;
    this._filter = input.value;
    this._open = true;
    this._highlightedIndex = -1;
    this.dispatchEvent(
      new CustomEvent('wf-input', {
        detail: { value: input.value },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleKeydown(e: KeyboardEvent): void {
    const filtered = this.filteredOptions;
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        this._open = true;
        this._highlightedIndex = Math.min(
          this._highlightedIndex + 1,
          filtered.length - 1,
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        this._highlightedIndex = Math.max(this._highlightedIndex - 1, 0);
        break;
      case 'Enter':
        e.preventDefault();
        if (
          this._open &&
          this._highlightedIndex >= 0 &&
          this._highlightedIndex < filtered.length
        ) {
          this._selectOption(filtered[this._highlightedIndex]);
        }
        break;
      case 'Escape':
        this.close();
        break;
    }
  }

  selectOption(option: { value: string; label: string }): void {
    this._selectOption(option);
  }

  private _selectOption(option: { value: string; label: string }): void {
    this.value = option.value;
    this._filter = '';
    this._open = false;
    this._highlightedIndex = -1;

    const input = this.querySelector(
      '.wf-combobox__input',
    ) as HTMLInputElement | null;
    if (input) input.value = option.label;

    this.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { value: option.value, label: option.label },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleOptionClick(option: { value: string; label: string }): void {
    this._selectOption(option);
  }

  private _selectedLabel(): string {
    const match = this.options.find((o) => o.value === this.value);
    return match ? match.label : '';
  }

  render() {
    const filtered = this.filteredOptions;
    const displayValue = this._filter || this._selectedLabel();

    return html`
      <label class="wf-field__label" ?hidden=${!this.label}>${this.label}</label>
      <div class="wf-field__container">
        <input
          type="text"
          class="wf-combobox__input wf-field__input"
          .value=${displayValue}
          placeholder=${this.placeholder || ''}
          ?disabled=${this.disabled}
          role="combobox"
          aria-expanded=${this._open}
          aria-autocomplete="list"
          @focus=${this._handleInputFocus}
          @input=${this._handleInput}
          @keydown=${this._handleKeydown}
        />
      </div>
      <ul class="wf-combobox__listbox" role="listbox">
        ${filtered.map(
          (opt, i) => html`
            <li
              class="wf-combobox__option ${i === this._highlightedIndex
                ? 'wf-combobox__option--highlighted'
                : ''} ${opt.value === this.value
                ? 'wf-combobox__option--selected'
                : ''}"
              role="option"
              aria-selected=${opt.value === this.value}
              @click=${() => this._handleOptionClick(opt)}
            >
              ${opt.label}
            </li>
          `,
        )}
      </ul>
    `;
  }
}

customElements.define('wf-combobox', WfCombobox);

declare global {
  interface HTMLElementTagNameMap {
    'wf-combobox': WfCombobox;
  }
}
