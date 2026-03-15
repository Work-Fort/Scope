import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-progress>` — Progress bar with determinate and indeterminate modes.
 *
 * Custom implementation instead of extending LionProgressIndicator because:
 * - LionProgressIndicator uses LocalizeMixin/getLocalizeManager for i18n default labels
 * - LionProgressIndicator's `_graphicTemplate()` returns `nothing` (expects subclassing with Shadow DOM)
 * - Our light DOM approach (createRenderRoot → this) is incompatible with Lion's localize infrastructure
 *
 * @element wf-progress
 */
export class WfProgress extends WfElement {
  @property({ type: Number }) value = 0;
  @property({ type: Number }) min = 0;
  @property({ type: Number }) max = 100;
  @property({ type: String }) size: 'sm' | 'md' | 'lg' = 'md';
  @property({ type: String }) label = '';
  @property({ type: Boolean }) indeterminate = false;

  private _container: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-progress');
    this.setAttribute('role', 'progressbar');
    this._ensureContainer();
    this._buildDOM();
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  /** Current progress as a percentage (0-100). */
  get percentage(): number {
    if (this.indeterminate) return 0;
    const clamped = Math.min(Math.max(this.value, this.min), this.max);
    return ((clamped - this.min) / (this.max - this.min)) * 100;
  }

  private _ensureContainer(): void {
    if (!this._container) {
      this._container = document.createElement('div');
      this._container.style.display = 'contents';
      this.appendChild(this._container);
    }
  }

  private _buildDOM(): void {
    if (!this._container) return;
    this._container.innerHTML = '';

    // Label row
    const labelRow = document.createElement('div');
    labelRow.className = 'wf-progress__label';

    const labelText = document.createElement('span');
    labelText.className = 'wf-progress__label-text';
    labelRow.appendChild(labelText);

    const valueText = document.createElement('span');
    valueText.className = 'wf-progress__value-text';
    labelRow.appendChild(valueText);

    this._container.appendChild(labelRow);

    // Track
    const track = document.createElement('div');
    track.className = 'wf-progress__track';

    const fill = document.createElement('div');
    fill.className = 'wf-progress__fill';
    track.appendChild(fill);

    this._container.appendChild(track);
  }

  private _sync(): void {
    // Size
    this.classList.remove('wf-progress--sm', 'wf-progress--md', 'wf-progress--lg');
    if (this.size !== 'md') this.classList.add(`wf-progress--${this.size}`);

    // Indeterminate
    this.classList.toggle('wf-progress--indeterminate', this.indeterminate);

    // ARIA
    if (this.indeterminate) {
      this.removeAttribute('aria-valuenow');
      this.removeAttribute('aria-valuemin');
      this.removeAttribute('aria-valuemax');
    } else {
      this.setAttribute('aria-valuenow', String(this.value));
      this.setAttribute('aria-valuemin', String(this.min));
      this.setAttribute('aria-valuemax', String(this.max));
    }

    // Fill width
    const fill = this.querySelector('.wf-progress__fill') as HTMLElement | null;
    if (fill && !this.indeterminate) {
      fill.style.width = `${this.percentage}%`;
    }

    // Label
    const labelEl = this.querySelector('.wf-progress__label-text');
    const labelRow = this.querySelector('.wf-progress__label') as HTMLElement | null;
    if (labelEl) labelEl.textContent = this.label;
    if (labelRow) {
      labelRow.style.display = this.label || !this.indeterminate ? '' : 'none';
    }

    // Value text
    const valueEl = this.querySelector('.wf-progress__value-text');
    if (valueEl) {
      valueEl.textContent = this.indeterminate ? '' : `${Math.round(this.percentage)}%`;
    }
  }
}

customElements.define('wf-progress', WfProgress);

declare global {
  interface HTMLElementTagNameMap {
    'wf-progress': WfProgress;
  }
}
