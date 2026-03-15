import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-step>` — A single step within a `<wf-stepper>`.
 *
 * @element wf-step
 */
export class WfStep extends WfElement {
  @property({ type: String }) label = '';
  @property({ type: String, reflect: true }) status: 'untouched' | 'active' | 'complete' = 'untouched';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-step');
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  private _sync(): void {
    this.classList.remove('wf-step--active', 'wf-step--complete', 'wf-step--untouched');
    this.classList.add(`wf-step--${this.status}`);
  }
}

customElements.define('wf-step', WfStep);

/**
 * `<wf-stepper>` — Multi-step progress indicator.
 *
 * Custom implementation instead of extending LionSteps because:
 * - LionSteps uses Shadow DOM with `<slot>` to find child steps via `shadowRoot.querySelector('slot')`
 * - LionStep uses Shadow DOM `:host` styles for show/hide based on status
 * - Our light DOM approach (createRenderRoot → this) is incompatible with slot-based child discovery
 *
 * @element wf-stepper
 * @fires wf-step-change - When the active step changes
 */
export class WfStepper extends WfElement {
  @property({ type: Number, reflect: true }) current = 0;
  @property({ type: String }) orientation: 'horizontal' | 'vertical' = 'horizontal';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-stepper');
    this._applyOrientation();
    this._buildIndicators();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('orientation')) this._applyOrientation();
    if (changed.has('current')) this._buildIndicators();
  }

  get steps(): WfStep[] {
    return Array.from(this.querySelectorAll('wf-step'));
  }

  next(): void {
    if (this.current < this.steps.length - 1) {
      this._goTo(this.current + 1);
    }
  }

  previous(): void {
    if (this.current > 0) {
      this._goTo(this.current - 1);
    }
  }

  goto(index: number): void {
    if (index >= 0 && index < this.steps.length) {
      this._goTo(index);
    }
  }

  private _goTo(index: number): void {
    const old = this.current;
    this.current = index;
    this._updateStepStatuses();
    this.dispatchEvent(
      new CustomEvent('wf-step-change', {
        bubbles: true,
        composed: true,
        detail: { from: old, to: index },
      }),
    );
  }

  private _updateStepStatuses(): void {
    this.steps.forEach((step, i) => {
      if (i < this.current) {
        step.status = 'complete';
      } else if (i === this.current) {
        step.status = 'active';
      } else {
        step.status = 'untouched';
      }
    });
  }

  private _applyOrientation(): void {
    this.classList.remove('wf-stepper--vertical', 'wf-stepper--horizontal');
    if (this.orientation === 'vertical') {
      this.classList.add('wf-stepper--vertical');
    }
  }

  private _buildIndicators(): void {
    // Remove existing indicators (but keep wf-step children)
    this.querySelectorAll('.wf-step__indicator, .wf-step__connector').forEach((el) =>
      el.remove(),
    );

    const steps = this.steps;
    this._updateStepStatuses();

    steps.forEach((step, i) => {
      // Prepend indicator to each step
      const existing = step.querySelector('.wf-step__indicator');
      if (!existing) {
        const indicator = document.createElement('span');
        indicator.className = 'wf-step__indicator';
        indicator.textContent = step.status === 'complete' ? '\u2713' : String(i + 1);
        step.prepend(indicator);
      } else {
        existing.textContent = step.status === 'complete' ? '\u2713' : String(i + 1);
      }

      // Update label
      const labelEl = step.querySelector('.wf-step__label');
      if (!labelEl && step.label) {
        const span = document.createElement('span');
        span.className = 'wf-step__label';
        span.textContent = step.label;
        step.appendChild(span);
      } else if (labelEl) {
        labelEl.textContent = step.label;
      }

      // Add connector between steps (not after last)
      if (i < steps.length - 1) {
        const nextSibling = step.nextElementSibling;
        if (!nextSibling?.classList.contains('wf-step__connector')) {
          const connector = document.createElement('span');
          connector.className = 'wf-step__connector';
          step.after(connector);
        }
      }
    });
  }
}

customElements.define('wf-stepper', WfStepper);

declare global {
  interface HTMLElementTagNameMap {
    'wf-step': WfStep;
    'wf-stepper': WfStepper;
  }
}
