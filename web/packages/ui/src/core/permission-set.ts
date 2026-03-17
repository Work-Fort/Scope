/**
 * Framework-agnostic permission checker.
 * Holds a set of permission strings and provides a can() method.
 * Framework adapters wrap this in reactive state.
 */
export class PermissionSet {
  private perms: Set<string>;

  constructor(permissions: string[]) {
    this.perms = new Set(permissions);
  }

  /** Check if a permission is granted. */
  can(permission: string): boolean {
    return this.perms.has(permission);
  }

  /** Replace all permissions. */
  update(permissions: string[]): void {
    this.perms = new Set(permissions);
  }

  /** Return all permissions as an array. */
  all(): string[] {
    return [...this.perms];
  }
}
