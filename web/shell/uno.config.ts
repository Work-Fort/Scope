import { defineConfig, presetWind } from 'unocss';

export default defineConfig({
  presets: [presetWind()],
  theme: {
    colors: {
      wf: {
        bg: 'var(--wf-bg)',
        'bg-secondary': 'var(--wf-bg-secondary)',
        text: 'var(--wf-text)',
        'text-secondary': 'var(--wf-text-secondary)',
        'text-muted': 'var(--wf-text-muted)',
        border: 'var(--wf-border)',
        accent: 'var(--wf-accent)',
        error: 'var(--wf-error)',
        warning: 'var(--wf-warning)',
        success: 'var(--wf-success)',
      },
    },
  },
});
