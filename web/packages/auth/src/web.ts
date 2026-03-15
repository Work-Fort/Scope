import type { AuthClient, UserInfo } from './index';

/**
 * Auth client for web environment (cookie-based sessions via Go BFF).
 * Relies on the browser's cookie jar — the Go BFF converts session cookies
 * to JWTs when proxying to backend services.
 */
export class WebAuthClient implements AuthClient {
  private listeners = new Set<(user: UserInfo | null) => void>();

  async getUser(): Promise<UserInfo | null> {
    try {
      const res = await fetch('/api/auth/me');
      if (!res.ok) return null;
      return res.json();
    } catch {
      return null;
    }
  }

  async login(email: string, password: string): Promise<UserInfo> {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      throw new Error(`Login failed: ${res.status}`);
    }
    const data = await res.json();
    this.notify(data.user ?? data);
    return data.user ?? data;
  }

  async logout(): Promise<void> {
    await fetch('/api/auth/logout', { method: 'POST' });
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
