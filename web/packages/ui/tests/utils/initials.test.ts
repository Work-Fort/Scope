import { describe, it, expect } from 'vitest';
import { initials } from '../../src/utils/initials.js';

describe('initials', () => {
  it('extracts two-part initials from hyphenated name', () => {
    expect(initials('alice-chen')).toBe('AC');
  });

  it('extracts two-part initials from underscored name', () => {
    expect(initials('bob_kim')).toBe('BK');
  });

  it('extracts first two chars for single-word name', () => {
    expect(initials('bob')).toBe('BO');
  });

  it('uppercases result', () => {
    expect(initials('alice-chen')).toBe('AC');
  });

  it('handles dotted names', () => {
    expect(initials('j.doe')).toBe('JD');
  });
});
