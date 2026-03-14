import { describe, it, expect } from 'vitest';
import {
  Validator,
  Required,
  IsString,
  EqualsLength,
  MinLength,
  MaxLength,
  MinMaxLength,
  IsEmail,
  Pattern,
  IsNumber,
  MinNumber,
  MaxNumber,
  MinMaxNumber,
  IsDate,
  MinDate,
  MaxDate,
  MinMaxDate,
} from '../../src/form/validators.js';

describe('Validator re-exports', () => {
  const validators = {
    Validator,
    Required,
    IsString,
    EqualsLength,
    MinLength,
    MaxLength,
    MinMaxLength,
    IsEmail,
    Pattern,
    IsNumber,
    MinNumber,
    MaxNumber,
    MinMaxNumber,
    IsDate,
    MinDate,
    MaxDate,
    MinMaxDate,
  };

  for (const [name, Ctor] of Object.entries(validators)) {
    it(`exports ${name}`, () => {
      expect(Ctor).toBeDefined();
      expect(typeof Ctor).toBe('function');
    });
  }

  it('Required is instantiable', () => {
    const v = new Required();
    expect(v).toBeInstanceOf(Validator);
  });

  it('MinLength is instantiable with param', () => {
    const v = new MinLength(3);
    expect(v).toBeInstanceOf(Validator);
  });
});
