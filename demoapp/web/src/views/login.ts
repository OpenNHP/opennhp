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
        <!-- OIDC RP not configured for this deployment; the handler returns
             503. Uncomment when an [[OIDC]] block is enabled in config.toml.
        <a id="login-oidc" class="btn btn-secondary" href="/api/auth/oidc/login">Sign in with OIDC</a>
        -->
        <a id="login-github" class="btn btn-secondary" href="/api/auth/github/login">
          <svg class="oauth-icon" viewBox="0 0 16 16" width="16" height="16" fill="currentColor" aria-hidden="true">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.02.37-2.45-.49-2.6-.67-.09-.23-.48-.67-.82-.82-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
          </svg>
          Sign in with GitHub
        </a>
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
