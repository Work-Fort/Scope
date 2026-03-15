import type { AuthClient, UserInfo } from './index';

/**
 * Auth client for Tauri environment.
 * Calls Tauri IPC commands instead of using cookies/fetch.
 * The Rust backend holds JWTs in memory.
 */
export class TauriAuthClient implements AuthClient {
  private listeners = new Set<(user: UserInfo | null) => void>();

  private async invoke<T>(cmd: string, args?: Record<string, unknown>): Promise<T> {
    // Use dynamic import to avoid bundling @tauri-apps/api in web builds.
    const { invoke } = await import('@tauri-apps/api/core');
    return invoke<T>(cmd, args);
  }

  async getUser(): Promise<UserInfo | null> {
    try {
      const user = await this.invoke<UserInfo | null>('get_user');
      return user;
    } catch {
      return null;
    }
  }

  async login(email: string, password: string): Promise<UserInfo> {
    const user = await this.invoke<UserInfo>('login', { email, password });
    this.notify(user);
    return user;
  }

  async logout(): Promise<void> {
    await this.invoke('logout');
    this.notify(null);
  }

  onAuthChange(callback: (user: UserInfo | null) => void): () => void {
    this.listeners.add(callback);
    return () => this.listeners.delete(callback);
  }

  private notify(user: UserInfo | null): void {
    for (const cb of this.listeners) {
      cb(user);
    }
  }
}
