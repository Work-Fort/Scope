import { describe, it, expect } from 'vitest';
import { formatTime, formatDateLabel, isSameDay } from '../../src/utils/time.js';

describe('formatTime', () => {
  it('formats ISO to HH:MM', () => {
    const result = formatTime('2026-03-15T09:14:00Z');
    expect(result).toMatch(/\d{2}:\d{2}/);
  });
});

describe('formatDateLabel', () => {
  it('returns date string for old dates', () => {
    const result = formatDateLabel('2020-01-01T12:00:00Z');
    expect(result).toContain('2020');
  });
});

describe('isSameDay', () => {
  it('returns true for same day', () => {
    expect(isSameDay('2026-03-15T09:00:00Z', '2026-03-15T23:00:00Z')).toBe(true);
  });

  it('returns false for different days', () => {
    expect(isSameDay('2026-03-13T12:00:00Z', '2026-03-15T12:00:00Z')).toBe(false);
  });
});
