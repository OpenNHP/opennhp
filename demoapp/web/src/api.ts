// Thin fetch wrapper for the demo backend. All methods share the browser's
// session cookie (the backend uses gin-contrib/sessions/cookie), so callers
// don't need to thread tokens through.
//
// Errors are normalized: HTTP non-2xx responses throw an ApiError so views
// can `try { await api.foo() } catch (e) { ... }` without inspecting Response.

export interface NhpEndpointConfig {
  serviceId: string;
  serverPublicKey: string;
  relayUrl: string;
  cipherScheme: 'curve25519' | 'gmsm';
  userId: string;
  organizationId: string;
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
  relayUrl: string;
  cipherScheme: 'curve25519' | 'gmsm';
  organizationId: string;
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

  register: (username: string, password: string, email: string) =>
    jsonFetch<RegisterResponse>('/api/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, email }),
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

  me: () => jsonFetch<{ username: string; email: string; status: string }>('/api/me'),

  credentials: () => jsonFetch<CredentialsResponse>('/api/credentials'),

  config: () => jsonFetch<ConfigResponse>('/api/config'),

  oidcLoginUrl: () => '/api/auth/oidc/login',
};
