// src/form/wf-textarea.ts
import { LionTextarea } from '@lion/ui/textarea.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * `<wf-textarea>` — Multi-line text input with auto-resize, validation,
 * and interaction states.
 * Extends LionTextarea with light DOM rendering and WorkFort CSS classes.
 *
 * @element wf-textarea
 */
export class WfTextarea extends LionTextarea {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    applyFieldHostClasses(this as any, 'wf-textarea');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}

customElements.define('wf-textarea', WfTextarea);

declare global {
  interface HTMLElementTagNameMap {
    'wf-textarea': WfTextarea;
  }
}
