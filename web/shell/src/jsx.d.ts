import 'solid-js';

declare module 'solid-js' {
  namespace JSX {
    interface IntrinsicElements {
      'wf-error-fallback': { title?: string; message?: string; class?: string };
      'wf-button': { variant?: string; disabled?: boolean; class?: string; 'on:wf-click'?: (e: CustomEvent) => void };
      'wf-list': { class?: string };
      'wf-list-item': { active?: boolean; class?: string; 'on:wf-select'?: (e: CustomEvent) => void };
      'wf-skeleton': { width?: string; height?: string; class?: string };
    }
  }
}
