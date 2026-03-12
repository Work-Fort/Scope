import { describe, it, expect, afterEach, vi } from 'vitest';
import React from 'react';
import { render, cleanup } from '@testing-library/react';
import '../../src/index.js';
import { Panel, Button, Badge } from '../../src/react/index.js';

describe('React component wrappers', () => {
  afterEach(cleanup);

  it('Panel renders wf-panel with label', () => {
    const { container } = render(<Panel label="Test">Content</Panel>);
    const panel = container.querySelector('wf-panel');
    expect(panel).not.toBeNull();
    expect(panel!.getAttribute('label')).toBe('Test');
  });

  it('Button renders wf-button with variant', () => {
    const { container } = render(<Button variant="filled">Click</Button>);
    const button = container.querySelector('wf-button');
    expect(button!.getAttribute('variant')).toBe('filled');
  });

  it('Badge renders wf-badge with count', () => {
    const { container } = render(<Badge count={5} />);
    const badge = container.querySelector('wf-badge');
    expect(badge!.getAttribute('count')).toBe('5');
  });

  it('forwards onWfClick event to addEventListener on the Custom Element', () => {
    const handler = vi.fn();
    const { container } = render(<Button onWfClick={handler}>Click me</Button>);
    const button = container.querySelector('wf-button')!;
    button.dispatchEvent(new CustomEvent('wf-click', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
  });
});
