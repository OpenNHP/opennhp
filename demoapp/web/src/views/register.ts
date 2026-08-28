// Register view — collect username/password/email, POST /api/register,
// then drive the NHP registration flow (requestOtp → registerPublicKey →
// confirm) via the shared NHP reg panel.

import { api, ApiError, type RegisterResponse } from '../api.js';
import { mountNhpRegPanel } from '../nhp-reg-panel.js';

export interface RegisterViewProps {
  onRegistered: () => void;
  onSwitchToLogin: () => void;
}

export function renderRegister(root: HTMLElement, props: RegisterViewProps): void {
  root.innerHTML = `
    <div class="container">
      <h1>OpenNHP Demo App</h1>
      <p class="subtitle">Create an account, then complete NHP registration to access protected resources.</p>

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

  function showAlert(level: 'error' | 'success' | 'info', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escape(message)}</div>`;
  }

  function escape(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  let disposePanel: (() => void) | undefined;

  submitBtn.addEventListener('click', async () => {
    const username = (root.querySelector<HTMLInputElement>('#reg-username')!).value.trim();
    const email = (root.querySelector<HTMLInputElement>('#reg-email')!).value.trim();
    const password = (root.querySelector<HTMLInputElement>('#reg-password')!).value;
    if (!username || !email || !password) {
      showAlert('error', 'All fields are required.');
      return;
    }
    submitBtn.disabled = true;
    showAlert('info', 'Creating account and generating NHP key pair…');
    let reg: RegisterResponse;
    try {
      reg = await api.register(username, password, email);
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
      onBack: () => {
        disposePanel?.();
        disposePanel = undefined;
        otpPanel.innerHTML = '';
        showAlert('info', 'Account created. Click "Create account & request OTP" again to retry.');
      },
    });
  });

  switchBtn.addEventListener('click', () => {
    disposePanel?.();
    props.onSwitchToLogin();
  });
}
