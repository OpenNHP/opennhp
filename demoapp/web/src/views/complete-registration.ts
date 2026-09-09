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
import { t, renderLangSwitcher } from '../i18n.js';
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
        <div class="user">${t('common.signedInAs')} <span>${escapeHtml(props.username)}</span></div>
        <button id="signout-btn" class="btn btn-secondary">${t('common.signOut')}</button>
      </div>
      <h1>${t('cr.title')}</h1>
      <p class="subtitle">${t('cr.subtitle')}</p>

      <div id="alert"></div>
      <div id="bind-area">
        <p class="note">${t('common.loading')}</p>
      </div>
      <div id="reg-area"></div>
    </div>
  `;

  renderLangSwitcher(root.querySelector<HTMLElement>('.container')!);

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
      showAlert('error', t('cr.loadCatalogFailed', { msg }));
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
      bindArea.innerHTML = `<div class="alert alert-error">${t('cr.noClusters')}</div>`;
      return;
    }
    // Default the selection to the current binding, else the first server.
    const selected = servers.find((s) => s.name === curServer) ?? servers[0];
    const selName = selected.name;

    bindArea.innerHTML = `
      <div class="panel">
        <h2>${t('cr.binding')}</h2>
        <div class="field">
          <label for="cr-server">${t('cr.cluster')}</label>
          <select id="cr-server">
            ${servers.map((s) => `<option value="${escapeHtml(s.name)}" ${s.name === selName ? 'selected' : ''}>${escapeHtml(s.name)}</option>`).join('')}
          </select>
        </div>
        <div class="field">
          <label for="cr-scheme">${t('cr.cipherScheme')}</label>
          <select id="cr-scheme"></select>
        </div>
        <button id="cr-confirm" class="btn btn-primary">${t('cr.confirm')}</button>
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
        showAlert('error', t('cr.selectClusterScheme'));
        return;
      }
      confirmBtn.disabled = true;
      showAlert('info', t('cr.generating'));
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
            showAlert('info', t('cr.clickRetry'));
          },
        });
      } catch (err) {
        const msg = err instanceof ApiError ? err.message : String(err);
        showAlert('error', t('cr.bindingFailed', { msg }));
      } finally {
        confirmBtn.disabled = false;
      }
    });
  }
}
