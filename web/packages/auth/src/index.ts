export interface UserInfo {
  id: string;
  email: string;
  name: string;
}

export interface AuthClient {
  getUser(): Promise<UserInfo | null>;
  login(email: string, password: string): Promise<UserInfo>;
  logout(): Promise<void>;
  onAuthChange(callback: (user: UserInfo | null) => void): () => void;
}

export { TauriAuthClient } from './tauri';
export { WebAuthClient } from './web';

/**
 * Factory: returns the correct auth client for the current environment.
 * Detects Tauri by checking for `window.__TAURI_INTERNALS__`.
 */
let _client: AuthClient | undefined;

export async function getAuthClient(): Promise<AuthClient> {
  if (_client) return _client;

  if (
    typeof window !== 'undefined' &&
    (window as any).__TAURI_INTERNALS__
  ) {
    const { TauriAuthClient } = await import('./tauri');
    _client = new TauriAuthClient();
  } else {
    const { WebAuthClient } = await import('./web');
    _client = new WebAuthClient();
  }

  return _client;
}
