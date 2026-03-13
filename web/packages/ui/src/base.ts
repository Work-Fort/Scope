import { LitElement } from 'lit';

/**
 * Base class for all @workfort/ui components.
 * Renders in light DOM (no Shadow DOM).
 */
export class WfElement extends LitElement {
  createRenderRoot(): this {
    return this;
  }
}
