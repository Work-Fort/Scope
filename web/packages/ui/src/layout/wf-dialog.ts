// wf-dialog is an alias for wf-modal
import { WfModal } from './wf-modal.js';

export { WfModal as WfDialog };

if (!customElements.get('wf-dialog')) {
  customElements.define('wf-dialog', class extends WfModal {});
}

declare global {
  interface HTMLElementTagNameMap {
    'wf-dialog': WfModal;
  }
}
