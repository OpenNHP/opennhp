// Complete-registration view — for users who created an account but never
// finished the NHP_REG handshake (status=pending). They're already logged
// in, so we fetch their current binding via /api/credentials and the server
// catalog via /api/servers, then let them (re)pick the nhp-server cluster +
// cipher scheme before driving the handshake. External-IdP (GitHub/OIDC)
// users land here with a default binding they never chose, so the chooser is
// essential; password users see their existing choice pre-selected.
//
// On confirm we POST /api/register/bind, which re-derives the public key
// under the chosen scheme (the private key is scheme-agnostic and is NOT
// rotated) and returns fresh reg material. We then mount the shared NHP reg
// panel, which runs requestOtp -> registerPublicKey -> confirm (session
// path, regToken omitted).

import { api, ApiError, type ServerInfo } from '../api.js';
import { escapeHtml } from '../escape.js';
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
        <div class="user">Signed in as <span>${escapeHtml(props.username)}</span></div>
        <button id="signout-btn" class="btn btn-secondary">Sign out</button>
      </div>
      <h1>Complete NHP Registration</h1>
      <p class="subtitle">Your account exists but the NHP key was never registered with nhp-server. Pick your cluster and cipher scheme, then complete the handshake to activate your account.</p>

      <div id="alert"></div>
      <div id="bind-area">
        <p class="note">Loading…</p>
      </div>
      <div id="reg-area"></div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const bindArea = root.querySelector<HTMLDivElement>('#bind-area')!;
  const regArea = root.querySelector<HTMLDivElement>('#reg-area')!;
  const signoutBtn = root.querySelector<HTMLButtonElement>('#signout-btn')!;

  function showAlert(level: 'error' | 'info' | 'success', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escapeHtml(message)}</div>`;
  }

  signoutBtn.addEventListener('click', async () => {
    await api.logout();
    props.onSignOut();
  });

  let disposePanel: (() => void) | undefined;

  void (async () => {
    let servers: ServerInfo[] = [];
    let curServer = '';
    let curScheme = '';
    try {
      const [srvRes, creds] = await Promise.all([api.servers(), api.credentials().catch(() => null)]);
      servers = srvRes.servers;
      if (creds) {
        curServer = creds.nhp.serverName;
        curScheme = creds.nhp.cipherScheme;
      }
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Failed to load server catalog: ${msg}`);
      return;
    }
    renderChooser(servers, curServer, curScheme);
  })();

  function renderChooser(
    servers: ServerInfo[],
    curServer: string,
    curScheme: string,
  ): void {
    if (servers.length === 0) {
      bindArea.innerHTML = `<div class="alert alert-error">No nhp-server clusters are configured.</div>`;
      return;
    }
    // Default the selection to the current binding, else the first server.
    const selected = servers.find((s) => s.name === curServer) ?? servers[0];
    const selName = selected.name;

    bindArea.innerHTML = `
      <div class="panel">
        <h2>NHP binding</h2>
        <div class="field">
          <label for="cr-server">NHP server cluster</label>
          <select id="cr-server">
            ${servers.map((s) => `<option value="${escapeHtml(s.name)}" ${s.name === selName ? 'selected' : ''}>${escapeHtml(s.name)}</option>`).join('')}
          </select>
        </div>
        <div class="field">
          <label for="cr-scheme">Cipher scheme</label>
          <select id="cr-scheme"></select>
        </div>
        <button id="cr-confirm" class="btn btn-primary">Confirm &amp; request OTP</button>
      </div>
    `;

    const serverSelect = bindArea.querySelector<HTMLSelectElement>('#cr-server')!;
    const schemeSelect = bindArea.querySelector<HTMLSelectElement>('#cr-scheme')!;
    const confirmBtn = bindArea.querySelector<HTMLButtonElement>('#cr-confirm')!;

    function syncSchemes(): void {
      const srv = servers.find((s) => s.name === serverSelect.value);
      if (!srv) {
        schemeSelect.innerHTML = `<option value="">—</option>`;
        return;
      }
      const schemes = srv.schemes.length > 0 ? srv.schemes : [srv.relayRegisteredScheme];
      const preferred = serverSelect.value === curServer && curScheme ? curScheme : srv.relayRegisteredScheme;
      schemeSelect.innerHTML = schemes
        .map((sc) => `<option value="${escapeHtml(sc)}" ${sc === preferred ? 'selected' : ''}>${escapeHtml(sc)}</option>`)
        .join('');
    }
    serverSelect.addEventListener('change', syncSchemes);
    syncSchemes();

    confirmBtn.addEventListener('click', async () => {
      const serverName = serverSelect.value;
      const cipherScheme = schemeSelect.value;
      if (!serverName || !cipherScheme) {
        showAlert('error', 'Select an NHP server cluster and cipher scheme.');
        return;
      }
      confirmBtn.disabled = true;
      showAlert('info', 'Generating NHP key material under the selected binding…');
      try {
        const reg = await api.registerBind(serverName, cipherScheme);
        // Hide the chooser once the panel is mounted.
        bindArea.innerHTML = '';
        disposePanel?.();
        disposePanel = mountNhpRegPanel(regArea, {
          privateKey: reg.privateKey,
          deviceId: reg.deviceId,
          email: reg.nhp.userId,
          nhp: reg.nhp,
          // regToken intentionally omitted — confirm uses the logged-in session.
          onComplete: (rakOk) => {
            if (rakOk) setTimeout(() => props.onCompleted(), 600);
          },
          onBack: () => {
            disposePanel?.();
            disposePanel = undefined;
            regArea.innerHTML = '';
            renderChooser(servers, serverName, cipherScheme);
            showAlert('info', 'Click "Confirm & request OTP" to retry.');
          },
        });
      } catch (err) {
        const msg = err instanceof ApiError ? err.message : String(err);
        showAlert('error', `Binding failed: ${msg}`);
      } finally {
        confirmBtn.disabled = false;
      }
    });
  }
}
