// src/form/validators.ts
// Re-export Lion validators so consumers import from @workfort/ui
// instead of directly from @lion/ui/form-core.js.

// @ts-expect-error — @lion/ui has no bundled type declarations
export { Validator } from '@lion/ui/form-core.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
export { Required } from '@lion/ui/form-core.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
export {
  IsString,
  EqualsLength,
  MinLength,
  MaxLength,
  MinMaxLength,
  IsEmail,
  Pattern,
} from '@lion/ui/form-core.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
export {
  IsNumber,
  MinNumber,
  MaxNumber,
  MinMaxNumber,
} from '@lion/ui/form-core.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
export {
  IsDate,
  MinDate,
  MaxDate,
  MinMaxDate,
} from '@lion/ui/form-core.js';
