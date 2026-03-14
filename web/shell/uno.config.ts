import { defineConfig, presetWind } from 'unocss';

export default defineConfig({
  presets: [presetWind()],
  theme: {
    colors: {
      wf: {
        bg: 'var(--wf-color-bg)',
        'bg-secondary': 'var(--wf-color-bg-secondary)',
        'bg-elevated': 'var(--wf-color-bg-elevated)',
        text: 'var(--wf-color-text)',
        'text-secondary': 'var(--wf-color-text-secondary)',
        'text-muted': 'var(--wf-color-text-muted)',
        'text-disabled': 'var(--wf-color-text-disabled)',
        border: 'var(--wf-color-border)',
        'border-focus': 'var(--wf-color-border-focus)',
        accent: 'var(--wf-color-accent)',
        error: 'var(--wf-color-error)',
        warning: 'var(--wf-color-warning)',
        success: 'var(--wf-color-success)',
        info: 'var(--wf-color-info)',
      },
    },
  },
});
