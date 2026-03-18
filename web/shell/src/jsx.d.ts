import 'solid-js';

declare module 'solid-js' {
  namespace JSX {
    interface IntrinsicElements {
      'wf-error-fallback': { title?: string; message?: string; class?: string; children?: JSX.Element };
      'wf-button': { variant?: string; disabled?: boolean; class?: string; 'on:wf-click'?: (e: CustomEvent) => void; children?: JSX.Element };
      'wf-list': { class?: string; children?: JSX.Element };
      'wf-list-item': { active?: boolean; class?: string; 'on:wf-select'?: (e: CustomEvent) => void; children?: JSX.Element };
      'wf-skeleton': { width?: string; height?: string; class?: string };
      'wf-nav-bar': { 'hamburger-position'?: string; class?: string; children?: JSX.Element };
      'wf-divider': { class?: string };
      'wf-status-dot': { status?: string; class?: string };
      'wf-banner': { variant?: string; headline?: string; class?: string; children?: JSX.Element };
    }
  }
}
