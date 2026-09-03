// Register view — collect username/password/email, pick a nhp-server and
// cipher scheme, POST /api/register, then drive the NHP registration flow
// (requestOtp → registerPublicKey → confirm) via the shared NHP reg panel.

import { api, ApiError, type RegisterResponse, type ServerInfo } from '../api.js';
import { escapeHtml } from '../escape.js';
import { mountNhpRegPanel } from '../nhp-reg-panel.js';

export interface RegisterViewProps {
  onRegistered: () => void;
  onSwitchToLogin: () => void;
}

export function renderRegister(root: HTMLElement, props: RegisterViewProps): void {
  root.innerHTML = `
    <div class="container">
      <h1>OpenNHP Integration Demo</h1>
      <p class="subtitle">A working example of adding the OpenNHP to an existing web application</p>

      <div id="alert"></div>

      <div class="panel">
        <h2>1. Account credentials</h2>
        <div class="field">
          <label for="reg-username">Username</label>
          <input id="reg-username" type="text" autocomplete="username" />
        </div>
        <div class="field">
          <label for="reg-email">Email</label>
          <input id="reg-email" type="email" autocomplete="email" />
        </div>
        <div class="field">
          <label for="reg-password">Password (min 8 chars)</label>
          <input id="reg-password" type="password" autocomplete="new-password" />
        </div>
        <div class="field">
          <label for="reg-server">NHP server</label>
          <select id="reg-server" disabled></select>
        </div>
        <div class="field">
          <label for="reg-scheme">Cipher scheme</label>
          <select id="reg-scheme" disabled>
            <option value="">Loading…</option>
          </select>
        </div>
        <button id="reg-submit" class="btn btn-primary">Create account &amp; request OTP</button>
        <button id="reg-switch-login" class="btn btn-secondary">Already have an account? Sign in</button>
      </div>

      <div id="otp-panel"></div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const submitBtn = root.querySelector<HTMLButtonElement>('#reg-submit')!;
  const switchBtn = root.querySelector<HTMLButtonElement>('#reg-switch-login')!;
  const otpPanel = root.querySelector<HTMLDivElement>('#otp-panel')!;
  const serverSelect = root.querySelector<HTMLSelectElement>('#reg-server')!;
  const schemeSelect = root.querySelector<HTMLSelectElement>('#reg-scheme')!;

  function showAlert(level: 'error' | 'success' | 'info', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escapeHtml(message)}</div>`;
  }

  // Populate the server + scheme dropdowns from GET /api/servers.
  let servers: ServerInfo[] = [];
  void (async () => {
    try {
      const res = await api.servers();
      servers = res.servers;
    } catch {
      servers = [];
    }
    if (servers.length === 0) {
      serverSelect.innerHTML = `<option value="">No servers configured</option>`;
      return;
    }
    serverSelect.innerHTML = servers
      .map((s) => `<option value="${escapeHtml(s.name)}">${escapeHtml(s.name)}</option>`)
      .join('');
    serverSelect.disabled = false;
    syncSchemeOptions();
  })();

  // Rebuild the scheme dropdown from the selected server's available
  // schemes, defaulting to its relay-registered scheme.
  function syncSchemeOptions(): void {
    const srv = servers.find((s) => s.name === serverSelect.value);
    if (!srv) {
      schemeSelect.innerHTML = `<option value="">—</option>`;
      return;
    }
    const schemes = srv.schemes.length > 0 ? srv.schemes : [srv.relayRegisteredScheme];
    schemeSelect.innerHTML = schemes
      .map((sc) => `<option value="${escapeHtml(sc)}" ${sc === srv.relayRegisteredScheme ? 'selected' : ''}>${escapeHtml(sc)}</option>`)
      .join('');
    schemeSelect.disabled = schemes.length === 0;
  }
  serverSelect.addEventListener('change', syncSchemeOptions);

  let disposePanel: (() => void) | undefined;

  submitBtn.addEventListener('click', async () => {
    const username = (root.querySelector<HTMLInputElement>('#reg-username')!).value.trim();
    const email = (root.querySelector<HTMLInputElement>('#reg-email')!).value.trim();
    const password = (root.querySelector<HTMLInputElement>('#reg-password')!.value);
    const serverName = serverSelect.value;
    const cipherScheme = schemeSelect.value;
    if (!username || !email || !password) {
      showAlert('error', 'All fields are required.');
      return;
    }
    if (!serverName || !cipherScheme) {
      showAlert('error', 'Select an NHP server and cipher scheme.');
      return;
    }
    submitBtn.disabled = true;
    showAlert('info', 'Creating account and generating NHP key pair…');
    let reg: RegisterResponse;
    try {
      reg = await api.register(username, password, email, serverName, cipherScheme);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Registration failed: ${msg}`);
      return;
    } finally {
      submitBtn.disabled = false;
    }

    // Account created; mount the NHP reg panel, which immediately wraps
    // the NHP-OTP packet and sends it via the relay to nhp-server.
    disposePanel?.();
    disposePanel = mountNhpRegPanel(otpPanel, {
      privateKey: reg.privateKey,
      deviceId: reg.deviceId,
      email: reg.nhp.userId,
      nhp: reg.nhp,
      regToken: reg.regToken,
      onComplete: (rakOk) => {
        if (rakOk) setTimeout(() => props.onRegistered(), 600);
      },
      // The account already exists, so re-POSTing /api/register would 409
      // (review #6). The credentials from the first /api/register call
      // are still in memory here, so "Back" just re-mounts the reg panel
      // with the same material and re-drives requestOtp — no backend
      // call, no 409. (If the user refreshed instead, the pending
      // session routes them to the complete-registration view, which
      // recovers via the session-gated /api/register/bind.) The retry
      // panel drops its own onBack so the user cannot loop the re-mount.
      onBack: () => {
        disposePanel?.();
        disposePanel = mountNhpRegPanel(otpPanel, {
          privateKey: reg.privateKey,
          deviceId: reg.deviceId,
          email: reg.nhp.userId,
          nhp: reg.nhp,
          regToken: reg.regToken,
          onComplete: (rakOk) => {
            if (rakOk) setTimeout(() => props.onRegistered(), 600);
          },
        });
      },
    });
  });

  switchBtn.addEventListener('click', () => {
    disposePanel?.();
    props.onSwitchToLogin();
  });
}
