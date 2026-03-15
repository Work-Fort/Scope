import { describe, it, expect } from 'vitest';
import '../../src/index.js';

const EXPECTED = [
  'wf-panel',
  'wf-button',
  'wf-badge',
  'wf-status-dot',
  'wf-skeleton',
  'wf-divider',
  'wf-text-input',
  'wf-list',
  'wf-list-item',
  'wf-scroll-area',
  'wf-error-fallback',
  'wf-input',
  'wf-textarea',
  'wf-select',
  'wf-checkbox',
  'wf-checkbox-group',
  'wf-radio',
  'wf-radio-group',
  'wf-toggle',
  'wf-slider',
  'wf-file-upload',
  'wf-combobox',
  'wf-date-picker',
  'wf-form',
  'wf-card',
  'wf-tabs',
  'wf-tab-panel',
  'wf-accordion',
  'wf-accordion-item',
  'wf-dialog',
  'wf-drawer',
  'wf-tooltip',
  'wf-popover',
  'wf-table',
];

describe('@workfort/ui registration', () => {
  it('registers all custom elements', () => {
    for (const tag of EXPECTED) {
      expect(customElements.get(tag), `${tag} should be registered`).toBeDefined();
    }
  });
});
