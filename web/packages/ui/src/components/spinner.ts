import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-spinner>` — Loading spinner indicator.
 * Renders an animated SVG spinner with configurable size.
 *
 * @element wf-spinner
 */
export class WfSpinner extends WfElement {
  @property({ type: String }) size: 'sm' | 'md' | 'lg' = 'md';
  @property({ type: String }) label = 'Loading';

  private _container: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-spinner');
    this.setAttribute('role', 'status');
    this._ensureContainer();
    this._buildSpinner();
  }

  updated(): void {
    this._applySize();
    this._updateLabel();
  }

  private _ensureContainer(): void {
    if (!this._container) {
      this._container = document.createElement('div');
      this._container.style.display = 'contents';
      this.appendChild(this._container);
    }
  }

  private _buildSpinner(): void {
    if (!this._container) return;
    this._container.innerHTML = '';

    // SVG spinner
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.classList.add('wf-spinner__svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('aria-hidden', 'true');

    const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    circle.classList.add('wf-spinner__track');
    circle.setAttribute('cx', '12');
    circle.setAttribute('cy', '12');
    circle.setAttribute('r', '10');
    circle.setAttribute('stroke', 'currentColor');
    circle.setAttribute('stroke-width', '2.5');
    svg.appendChild(circle);

    const arc = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    arc.classList.add('wf-spinner__arc');
    arc.setAttribute('cx', '12');
    arc.setAttribute('cy', '12');
    arc.setAttribute('r', '10');
    arc.setAttribute('stroke', 'currentColor');
    arc.setAttribute('stroke-width', '2.5');
    arc.setAttribute('stroke-linecap', 'round');
    arc.setAttribute('stroke-dasharray', '40 60');
    svg.appendChild(arc);

    this._container.appendChild(svg);

    // Screen reader text
    const srText = document.createElement('span');
    srText.className = 'wf-spinner__label';
    srText.textContent = this.label;
    this._container.appendChild(srText);

    this._applySize();
  }

  private _applySize(): void {
    this.classList.remove('wf-spinner--sm', 'wf-spinner--md', 'wf-spinner--lg');
    this.classList.add(`wf-spinner--${this.size}`);
  }

  private _updateLabel(): void {
    const labelEl = this.querySelector('.wf-spinner__label');
    if (labelEl) labelEl.textContent = this.label;
  }
}

customElements.define('wf-spinner', WfSpinner);

declare global {
  interface HTMLElementTagNameMap {
    'wf-spinner': WfSpinner;
  }
}
