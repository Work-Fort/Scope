// src/form/wf-field.ts
import { LionField } from '@lion/ui/form-core.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * Base class for WorkFort form fields.
 * Extends LionField with light DOM rendering and WorkFort CSS classes.
 * Uses Lion's default slot-based rendering, applying WorkFort CSS classes
 * to the generated elements.
 *
 * Subclasses override:
 * - `_wfTag` for the component-specific CSS class (e.g., 'wf-input')
 */
export class WfField extends LionField {
  createRenderRoot(): this {
    return this;
  }

  /** CSS class added to host. Override in subclasses. */
  protected get _wfTag(): string {
    return '';
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field');
    if (this._wfTag) this.classList.add(this._wfTag);
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}
