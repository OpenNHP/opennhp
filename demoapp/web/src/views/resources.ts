// Resources view — main app screen. Fetches credentials, builds the
// NHPAgent, intersects the dynamic list from listServices with the
// backend's catalog, and renders a knock button per resource.

import { api, ApiError, type ConfigResponse, type ResourceMeta } from '../api.js';
import { createAgent, listResources, knockResource } from '../nhp.js';

export interface ResourcesViewProps {
  username: string;
  email: string;
  cipherScheme: string;
  serverName: string;
  authProvider: string;
  onSignOut: () => void;
}

// authProviderLabel turns the raw auth_provider value into a friendly
// badge label. Empty (legacy rows pre-dating the column) reads as local.
function authProviderLabel(p: string): string {
  switch (p) {
    case 'github': return 'GitHub';
    case 'oidc': return 'OIDC';
    case 'password': return 'Local';
    default: return 'Local';
  }
}

export function renderResources(root: HTMLElement, props: ResourcesViewProps): void {
  const scheme = props.cipherScheme || '—';
  const server = props.serverName || '—';
  const provider = authProviderLabel(props.authProvider);
  root.innerHTML = `
    <div class="container">
      <div class="toolbar">
        <div class="toolbar-id">
          <div class="user">Signed in as <span>${escape(props.username)}</span></div>
          <div class="badges">
            <span class="badge"><span class="badge-k">Auth</span>${escape(provider)}</span>
            <span class="badge badge-mono"><span class="badge-k">Alg</span>${escape(scheme)}</span>
            <span class="badge"><span class="badge-k">Server</span>${escape(server)}</span>
          </div>
        </div>
        <div class="toolbar-actions">
          <button id="signout-btn" class="btn btn-secondary">Sign out</button>
          <button id="delete-account-btn" class="btn btn-danger">Delete account</button>
        </div>
      </div>
      <h1>Protected Resources</h1>
      <p class="subtitle">Click "Access" to knock nhp-server. The protected resource is hidden until the knock opens the door.</p>

      <div id="alert"></div>
      <div id="resource-area">
        <p class="note">Loading…</p>
      </div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const area = root.querySelector<HTMLDivElement>('#resource-area')!;
  const signoutBtn = root.querySelector<HTMLButtonElement>('#signout-btn')!;
  const deleteAccountBtn = root.querySelector<HTMLButtonElement>('#delete-account-btn')!;

  function showAlert(level: 'error' | 'success' | 'info', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escape(message)}</div>`;
  }

  function escape(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  signoutBtn.addEventListener('click', async () => {
    await api.logout();
    props.onSignOut();
  });

  deleteAccountBtn.addEventListener('click', async () => {
    // Irreversible: the sealed NHP private key is deleted from the backend,
    // so the account (and its registered identity) cannot be recovered. The
    // nhp-server public key is left to expire via the server TTL.
    const ok = window.confirm(
      'Delete your account? This permanently removes your credentials and NHP ' +
      'key material from this demo. You can re-register afterward. This cannot be undone.',
    );
    if (!ok) return;
    try {
      await api.deleteAccount();
      props.onSignOut();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Failed to delete account: ${msg}`);
    }
  });

  void bootstrap();

  async function bootstrap(): Promise<void> {
    showAlert('info', 'Fetching credentials from server…');
    let cfg: ConfigResponse;
    let creds;
    try {
      const [c, cr] = await Promise.all([api.config(), api.credentials()]);
      cfg = c;
      creds = cr;
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', `Failed to load: ${msg}`);
      return;
    }

    showAlert('info', 'Initializing NHP agent and listing services…');
    const handle = await createAgent(creds.privateKey, creds.nhp, creds.deviceId, 'error');
    try {
      const ids = await listResources(handle, creds.nhp);
      handle.dispose();

      // Intersect with backend catalog; backend is the source of UI metadata.
      const byId = new Map<string, ResourceMeta>();
      cfg.resources.forEach((r) => byId.set(r.id, r));
      const visible = ids
        .map((id) => byId.get(id))
        .filter((r): r is ResourceMeta => Boolean(r));

      if (visible.length === 0) {
        area.innerHTML = `<p class="note">No accessible resources. Confirm registration completed and that the nhp-server basic plugin allows your user.</p>`;
        return;
      }

      alert.innerHTML = '';
      area.innerHTML = `
        <ul class="resource-list">
          ${visible.map((r) => `
            <li class="resource-item" data-resid="${escape(r.id)}" data-url="${escape(r.url)}">
              <div>
                <div class="resource-title">${escape(r.title)}</div>
                <div class="resource-id">id: ${escape(r.id)} &middot; ${escape(r.url)}</div>
              </div>
              <button class="btn btn-primary knock-btn" data-resid="${escape(r.id)}">Access</button>
            </li>
          `).join('')}
        </ul>
      `;
      area.querySelectorAll<HTMLButtonElement>('.knock-btn').forEach((btn) => {
        btn.addEventListener('click', () => {
          // Open the blank window synchronously inside the user gesture so
          // popup blockers don't suppress it; we set its location only after
          // the (async) knock succeeds.
          const popup = window.open('', '_blank');
          void doKnock(btn.dataset.resid || '', popup);
        });
      });
    } catch (err) {
      handle.dispose();
      const msg = err instanceof Error ? err.message : String(err);
      showAlert('error', `listServices failed: ${msg}`);
    }
  }

  async function doKnock(resourceId: string, popup: Window | null): Promise<void> {
    const item = area.querySelector<HTMLLIElement>(`li.resource-item[data-resid="${resourceId}"]`);
    const btn = item?.querySelector<HTMLButtonElement>('.knock-btn');
    if (btn) btn.disabled = true;
    showAlert('info', `Knocking ${resourceId}…`);
    try {
      const creds = await api.credentials();
      const handle = await createAgent(creds.privateKey, creds.nhp, creds.deviceId, 'error');
      const outcome = await knockResource(handle, creds.nhp, resourceId);
      handle.dispose();
      if (outcome.success && outcome.resourceHost) {
        showAlert('success', `Knock successful — opening ${outcome.resourceHost}`);
        const url = /^https?:\/\//.test(outcome.resourceHost)
          ? outcome.resourceHost
          : `https://${outcome.resourceHost}`;
        if (popup && !popup.closed) {
          popup.location.href = url;
        } else {
          // Popup was blocked or already closed — fall back to navigating the
          // current tab so the knock isn't wasted.
          window.location.href = url;
        }
      } else {
        showAlert('error', `Knock failed: ${outcome.error ?? 'unknown'}`);
        popup?.close();
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      showAlert('error', `Knock failed: ${msg}`);
      popup?.close();
    } finally {
      if (btn) btn.disabled = false;
    }
  }
}
