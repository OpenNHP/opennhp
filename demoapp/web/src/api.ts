// Thin fetch wrapper for the demo backend. All methods share the browser's
// session cookie (the backend uses gin-contrib/sessions/cookie), so callers
// don't need to thread tokens through.
//
// Errors are normalized: HTTP non-2xx responses throw an ApiError so views
// can `try { await api.foo() } catch (e) { ... }` without inspecting Response.

export interface NhpEndpointConfig {
  serviceId: string;
  serverPublicKey: string;
  /** Key the relay registered the server under; js-agent fingerprints it for
   * /relay/<id> routing. Empty when the chosen scheme matches the relay's
   * registered scheme (the ECDH key already produces the right fingerprint). */
  relayPublicKey?: string;
  relayUrl: string;
  cipherScheme: 'curve25519' | 'gmsm';
  userId: string;
  organizationId: string;
  serverName: string;
}

export interface ServerInfo {
  name: string;
  serviceId: string;
  organizationId: string;
  /** Schemes the server can actually serve (those whose public key is
   * configured). A docker server with only the SM2 key offers ['gmsm']. */
  schemes: string[];
  relayRegisteredScheme: 'curve25519' | 'gmsm';
}

export interface RegisterResponse {
  regToken: string;
  privateKey: string;
  publicKey: string;
  deviceId: string;
  nhp: NhpEndpointConfig;
}

export interface CredentialsResponse {
  privateKey: string;
  publicKey: string;
  deviceId: string;
  userId: string;
  nhp: NhpEndpointConfig;
}

export interface ResourceMeta {
  id: string;
  title: string;
  url: string;
  acHost: string;
}

export interface ConfigResponse {
  serviceId: string;
  serverPublicKey: string;
  relayPublicKey?: string;
  relayUrl: string;
  cipherScheme: 'curve25519' | 'gmsm';
  organizationId: string;
  serverName: string;
  userId: string;
  resources: ResourceMeta[];
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function jsonFetch<T>(input: string, init?: RequestInit): Promise<T> {
  const r = await fetch(input, {
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  });
  if (!r.ok) {
    let msg = `${r.status}`;
    try {
      const body = await r.json();
      if (body && typeof body.error === 'string') msg = body.error;
    } catch {
      // body wasn't JSON
    }
    throw new ApiError(msg, r.status);
  }
  // 204 No Content
  if (r.status === 204) return undefined as T;
  return r.json() as Promise<T>;
}

export const api = {
  health: () => jsonFetch<{ status: string }>('/api/health'),

  servers: () => jsonFetch<{ servers: ServerInfo[] }>('/api/servers'),

  register: (username: string, password: string, email: string, serverName = '', cipherScheme = '') =>
    jsonFetch<RegisterResponse>('/api/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, email, serverName, cipherScheme }),
    }),

  registerRetry: (regToken: string) =>
    jsonFetch<RegisterResponse>('/api/register/retry', {
      method: 'POST',
      body: JSON.stringify({ regToken }),
    }),

  registerConfirm: (regToken: string, deviceId: string, expiresAt: number, rakOk: boolean) =>
    jsonFetch<{ success: boolean }>('/api/register/confirm', {
      method: 'POST',
      body: JSON.stringify({ regToken, deviceId, expiresAt, rakOk }),
    }),

  login: (username: string, password: string) =>
    jsonFetch<{ username: string; email: string; status: string }>('/api/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  logout: () =>
    jsonFetch<{ success: boolean }>('/api/logout', {
      method: 'POST',
    }),

  /** Permanently delete the signed-in user's local account (credentials,
   * sealed NHP private key, deviceId). The session is cleared server-side.
   * The NHP-server public key is left to expire via the server TTL. */
  deleteAccount: () =>
    jsonFetch<{ success: boolean }>('/api/account', {
      method: 'DELETE',
    }),

  me: () => jsonFetch<{ username: string; email: string; status: string }>('/api/me'),

  credentials: () => jsonFetch<CredentialsResponse>('/api/credentials'),

  config: () => jsonFetch<ConfigResponse>('/api/config'),

  oidcLoginUrl: () => '/api/auth/oidc/login',

  githubLoginUrl: () => '/api/auth/github/login',
};
