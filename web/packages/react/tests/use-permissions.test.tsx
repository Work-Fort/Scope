import { describe, it, expect, afterEach } from 'vitest';
import React from 'react';
import { render, cleanup, act } from '@testing-library/react';
import { usePermissions } from '../src/use-permissions.js';

function TestComponent({ initial }: { initial?: string[] }) {
  const { can, update } = usePermissions(initial);
  return (
    <div>
      <span data-testid="send">{can('send_message') ? 'yes' : 'no'}</span>
      <span data-testid="manage">{can('manage_roles') ? 'yes' : 'no'}</span>
      <button data-testid="update" onClick={() => update(['manage_roles'])}>update</button>
    </div>
  );
}

describe('usePermissions (React)', () => {
  afterEach(() => { cleanup(); });

  it('checks initial permissions', () => {
    const { getByTestId } = render(<TestComponent initial={['send_message']} />);
    expect(getByTestId('send').textContent).toBe('yes');
    expect(getByTestId('manage').textContent).toBe('no');
  });

  it('starts with empty permissions by default', () => {
    const { getByTestId } = render(<TestComponent />);
    expect(getByTestId('send').textContent).toBe('no');
  });

  it('updates permissions and re-renders', () => {
    const { getByTestId } = render(<TestComponent initial={[]} />);
    expect(getByTestId('manage').textContent).toBe('no');
    act(() => { getByTestId('update').click(); });
    expect(getByTestId('manage').textContent).toBe('yes');
    expect(getByTestId('send').textContent).toBe('no');
  });
});
