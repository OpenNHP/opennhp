// Complete-registration view — for users who created an account but never
// finished the NHP_REG handshake (status=pending). They're already logged
// in, so we fetch their server-sealed credentials via /api/credentials and
// re-run the NHP registration (requestOtp → registerPublicKey → confirm)
// using the shared panel. The confirm call goes through the session-based
// path (regToken omitted).

import { api, ApiError } from '../api.js';
import { mountNhpRegPanel } from '../nhp-reg-panel.js';

export interface CompleteRegistrationViewProps {
  username: string;
  email: string;
  onCompleted: () => void;
  onSignOut: () => void;
}

export function renderCompleteRegistration(root: HTMLElement, props: CompleteRegistrationViewProps): void {
  root.innerHTML = `
    <div class="container">
      <div class="toolbar">
        <div class="user">Signed in as <span>${escape(props.username)}</span> (${escape(props.email)})</div>
        <button id="signout-btn" class="btn btn-secondary">Sign out</button>
      </div>
      <h1>Complete NHP Registration</h1>
      <p class="subtitle">Your account exists but the NHP key was never registered with nhp-server. Complete the handshake below to activate your account.</p>

      <div id="alert"></div>
      <div id="reg-area">
        <p class="note">Loading credentials…</p>
      </div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const area = root.querySelector<HTMLDivElement>('#reg-area')!;
  const signoutBtn = root.querySelector<HTMLButtonElement>('#signout-btn')!;

  function showAlert(level: 'error' | 'info' | 'success', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escape(message)}</div>`;
  }

  function escape(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  signoutBtn.addEventListener('click', async () => {
    await api.logout();
    props.onSignOut();
  });

  let disposePanel: (() => void) | undefined;

  void (async () => {
    let creds;
    try {
      creds = await api.credentials();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Failed to load credentials: ${msg}`);
      return;
    }
    if (!creds?.privateKey || !creds?.deviceId || !creds?.nhp) {
      showAlert('error', 'No NHP credentials on file. Please re-register from scratch.');
      return;
    }

    disposePanel?.();
    disposePanel = mountNhpRegPanel(area, {
      privateKey: creds.privateKey,
      deviceId: creds.deviceId,
      email: creds.userId,
      nhp: creds.nhp,
      // regToken intentionally omitted — confirm uses the logged-in session.
      onComplete: (rakOk) => {
        if (rakOk) setTimeout(() => props.onCompleted(), 600);
      },
      onBack: () => {
        disposePanel?.();
        disposePanel = undefined;
        area.innerHTML = '<p class="note">OTP request cancelled.</p>';
        showAlert('info', 'Click sign out or refresh to retry.');
      },
    });
  })();
}
