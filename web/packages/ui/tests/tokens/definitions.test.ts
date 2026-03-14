// tests/tokens/definitions.test.ts
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { describe, it, expect } from 'vitest';

const __dirname = dirname(fileURLToPath(import.meta.url));
const stylesDir = resolve(__dirname, '../../src/styles');

function readCSS(filename: string): string {
  return readFileSync(resolve(stylesDir, filename), 'utf8');
}

describe('primitives.css', () => {
  const css = readCSS('primitives.css');

  it('defines stone palette (50-950)', () => {
    for (const stop of [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950]) {
      expect(css, `missing --wf-stone-${stop}`).toContain(`--wf-stone-${stop}`);
    }
  });

  it('defines status color palettes', () => {
    for (const hue of ['red', 'amber', 'green', 'blue']) {
      for (const stop of [50, 500, 950]) {
        expect(css, `missing --wf-${hue}-${stop}`).toContain(`--wf-${hue}-${stop}`);
      }
    }
  });

  it('defines numeric spacing scale', () => {
    for (const n of [1, 2, 3, 4, 5, 6, 8, 10, 12, 16, 20, 24]) {
      expect(css, `missing --wf-space-${n}`).toContain(`--wf-space-${n}:`);
    }
  });

  it('defines typography size primitives', () => {
    for (const size of ['xs', 'sm', 'base', 'md', 'lg', 'xl', '2xl', '3xl', '4xl']) {
      expect(css, `missing --wf-text-${size}`).toContain(`--wf-text-${size}`);
    }
  });

  it('defines font weight primitives', () => {
    for (const w of ['normal', 'medium', 'semibold', 'bold']) {
      expect(css, `missing --wf-weight-${w}`).toContain(`--wf-weight-${w}`);
    }
  });

  it('defines line-height primitives', () => {
    for (const lh of ['tight', 'normal', 'relaxed']) {
      expect(css, `missing --wf-leading-${lh}`).toContain(`--wf-leading-${lh}`);
    }
  });

  it('defines font stack primitives', () => {
    expect(css).toContain('--wf-font-sans');
    expect(css).toContain('--wf-font-mono');
  });

  it('defines radius primitives', () => {
    for (const r of ['none', 'xs', 'sm', 'md', 'lg', 'xl', 'full']) {
      expect(css, `missing --wf-radius-${r}`).toContain(`--wf-radius-${r}`);
    }
  });

  it('defines shadow primitives', () => {
    for (const s of ['sm', 'md', 'lg', 'xl']) {
      expect(css, `missing --wf-shadow-${s}`).toContain(`--wf-shadow-${s}`);
    }
  });

  it('defines z-index scale', () => {
    for (const z of ['dropdown', 'sticky', 'modal', 'toast', 'tooltip']) {
      expect(css, `missing --wf-z-${z}`).toContain(`--wf-z-${z}`);
    }
  });

  it('defines motion primitives', () => {
    for (const d of ['fast', 'normal', 'slow']) {
      expect(css, `missing --wf-duration-${d}`).toContain(`--wf-duration-${d}`);
    }
    for (const e of ['in', 'out', 'in-out']) {
      expect(css, `missing --wf-ease-${e}`).toContain(`--wf-ease-${e}`);
    }
  });
});

describe('tokens.css', () => {
  const css = readCSS('tokens.css');

  it('defines semantic color tokens', () => {
    const expected = [
      '--wf-color-bg', '--wf-color-bg-secondary', '--wf-color-bg-elevated',
      '--wf-color-text', '--wf-color-text-secondary', '--wf-color-text-muted',
      '--wf-color-border', '--wf-color-accent',
      '--wf-color-error', '--wf-color-warning', '--wf-color-success', '--wf-color-info',
    ];
    for (const token of expected) {
      expect(css, `missing ${token}`).toContain(token);
    }
  });

  it('references primitives via var() — no hardcoded hex', () => {
    expect(css).toContain('var(--wf-stone-');
    const lines = css.split('\n').filter(l =>
      l.includes('--wf-color-') && l.includes(':') && !l.trim().startsWith('/*')
    );
    const hexInDefinition = lines.filter(l => /#[0-9a-f]{6}\b/i.test(l));
    expect(hexInDefinition, 'found hardcoded hex in semantic token definitions').toHaveLength(0);
  });

  it('has dark and light themes', () => {
    expect(css).toContain(':root');
    expect(css).toContain('[data-theme="light"]');
  });

  it('defines named spacing aliases referencing numeric primitives', () => {
    for (const name of ['xs', 'sm', 'md', 'lg', 'xl']) {
      expect(css, `missing --wf-space-${name}`).toContain(`--wf-space-${name}`);
    }
    expect(css).toContain('var(--wf-space-');
  });
});

describe('no old token names in component styles', () => {
  const files = ['components.css', 'banner.css', 'toast.css'];

  for (const file of files) {
    describe(file, () => {
      const css = readCSS(file);

      it('uses --wf-color-* (not old --wf-bg/--wf-text color names)', () => {
        const oldColorRefs = [
          'var(--wf-bg)',
          'var(--wf-bg-secondary)',
          'var(--wf-text)',
          'var(--wf-text-secondary)',
          'var(--wf-text-muted)',
          'var(--wf-border)',
          'var(--wf-accent)',
          'var(--wf-error)',
          'var(--wf-error-subtle)',
          'var(--wf-warning)',
          'var(--wf-warning-subtle)',
          'var(--wf-success)',
          'var(--wf-success-subtle)',
        ];
        for (const ref of oldColorRefs) {
          expect(css, `${file} still uses old token: ${ref}`).not.toContain(ref);
        }
      });

      it('uses --wf-text-* (not old --wf-font-size-* names)', () => {
        expect(css, `${file} still uses old --wf-font-size-*`).not.toContain('var(--wf-font-size-');
      });
    });
  }
});
