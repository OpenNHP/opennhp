// Login view — username/password. OIDC is exposed as a single "Sign in
// with OIDC" button that just navigates to the backend's /api/auth/oidc/login.

import { api, ApiError } from '../api.js';

export interface LoginViewProps {
  onSignedIn: () => void;
  onSwitchToRegister: () => void;
}

export function renderLogin(root: HTMLElement, props: LoginViewProps): void {
  root.innerHTML = `
    <div class="container">
      <h1>OpenNHP Demo App</h1>
      <p class="subtitle">Sign in to access NHP-protected resources.</p>

      <div id="alert"></div>

      <div class="panel">
        <div class="field">
          <label for="login-username">Username</label>
          <input id="login-username" type="text" autocomplete="username" />
        </div>
        <div class="field">
          <label for="login-password">Password</label>
          <input id="login-password" type="password" autocomplete="current-password" />
        </div>
        <button id="login-submit" class="btn btn-primary">Sign in</button>
        <a id="login-oidc" class="btn btn-secondary" href="/api/auth/oidc/login">Sign in with OIDC</a>
        <a id="login-github" class="btn btn-secondary" href="/api/auth/github/login">Sign in with GitHub</a>
        <button id="login-switch-register" class="btn btn-secondary">New here? Create an account</button>
      </div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const submitBtn = root.querySelector<HTMLButtonElement>('#login-submit')!;
  const switchBtn = root.querySelector<HTMLButtonElement>('#login-switch-register')!;

  function showAlert(level: 'error' | 'success' | 'info', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escape(message)}</div>`;
  }

  function escape(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  submitBtn.addEventListener('click', async () => {
    const username = (root.querySelector<HTMLInputElement>('#login-username')!).value.trim();
    const password = (root.querySelector<HTMLInputElement>('#login-password')!).value;
    if (!username || !password) {
      showAlert('error', 'Username and password are required.');
      return;
    }
    submitBtn.disabled = true;
    showAlert('info', 'Signing in…');
    try {
      await api.login(username, password);
      props.onSignedIn();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Sign in failed: ${msg}`);
    } finally {
      submitBtn.disabled = false;
    }
  });

  switchBtn.addEventListener('click', () => props.onSwitchToRegister());
}
