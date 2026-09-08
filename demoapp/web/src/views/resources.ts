// Resources view — main app screen. Fetches credentials, builds the
// NHPAgent, intersects the dynamic list from listServices with the
// backend's catalog, and renders a knock button per resource.

import { api, ApiError, type ConfigResponse, type ResourceMeta } from '../api.js';
import { escapeHtml } from '../escape.js';
import { t } from '../i18n.js';
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
    case 'password': return t('res.providerLocal');
    default: return t('res.providerLocal');
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
          <div class="user">${t('common.signedInAs')} <span>${escapeHtml(props.username)}</span></div>
          <div class="badges">
            <span class="badge"><span class="badge-k">${t('res.badgeAuth')}</span>${escapeHtml(provider)}</span>
            <span class="badge badge-mono"><span class="badge-k">${t('res.badgeAlg')}</span>${escapeHtml(scheme)}</span>
            <span class="badge"><span class="badge-k">${t('res.badgeServer')}</span>${escapeHtml(server)}</span>
          </div>
        </div>
        <div class="toolbar-actions">
          <button id="signout-btn" class="btn btn-secondary">${t('common.signOut')}</button>
          <button id="delete-account-btn" class="btn btn-danger">${t('res.deleteAccount')}</button>
        </div>
      </div>
      <h1>${t('res.title')}</h1>
      <p class="subtitle">${t('res.subtitle')}</p>

      <div id="alert"></div>
      <div id="resource-area">
        <p class="note">${t('common.loading')}</p>
      </div>
    </div>
  `;

  const alert = root.querySelector<HTMLDivElement>('#alert')!;
  const area = root.querySelector<HTMLDivElement>('#resource-area')!;
  const signoutBtn = root.querySelector<HTMLButtonElement>('#signout-btn')!;
  const deleteAccountBtn = root.querySelector<HTMLButtonElement>('#delete-account-btn')!;

  function showAlert(level: 'error' | 'success' | 'info', message: string): void {
    alert.innerHTML = `<div class="alert alert-${level}">${escapeHtml(message)}</div>`;
  }

  signoutBtn.addEventListener('click', async () => {
    await api.logout();
    props.onSignOut();
  });

  deleteAccountBtn.addEventListener('click', async () => {
    // Irreversible: the sealed NHP private key is deleted from the backend,
    // so the account (and its registered identity) cannot be recovered. The
    // nhp-server public key is left to expire via the server TTL.
    const ok = window.confirm(t('res.deleteConfirm'));
    if (!ok) return;
    try {
      await api.deleteAccount();
      props.onSignOut();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', t('res.deleteFailed', { msg }));
    }
  });

  void bootstrap();

  async function bootstrap(): Promise<void> {
    showAlert('info', t('res.fetchingCreds'));
    let cfg: ConfigResponse;
    let creds;
    try {
      const [c, cr] = await Promise.all([api.config(), api.credentials()]);
      cfg = c;
      creds = cr;
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      showAlert('error', t('res.loadFailed', { msg }));
      return;
    }

    showAlert('info', t('res.initAgent'));
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
        area.innerHTML = `<p class="note">${t('res.none')}</p>`;
        return;
      }

      alert.innerHTML = '';
      area.innerHTML = `
        <ul class="resource-list">
          ${visible.map((r) => `
            <li class="resource-item" data-resid="${escapeHtml(r.id)}" data-url="${escapeHtml(r.url)}">
              <div>
                <div class="resource-title">${escapeHtml(r.title)}</div>
                <div class="resource-id">${t('res.idPrefix')}: ${escapeHtml(r.id)} &middot; ${escapeHtml(r.url)}</div>
              </div>
              <button class="btn btn-primary knock-btn" data-resid="${escapeHtml(r.id)}">${t('res.access')}</button>
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
      showAlert('error', t('res.listFailed', { msg }));
    }
  }

  async function doKnock(resourceId: string, popup: Window | null): Promise<void> {
    // Match the list item by walking the rendered DOM rather than via a
    // querySelector with resourceId interpolated into the selector string.
    // The Id is the operator-authored [[Resources]] Id and is escaped for
    // HTML attribute output, but querySelector is a separate parser — a " or
    // \ in the value would throw SyntaxError before this function can
    // wrap it in try/catch, leaving the button enabled and the already-
    // opened about:blank popup orphaned.
    let item: HTMLLIElement | undefined;
    for (const li of Array.from(area.querySelectorAll<HTMLLIElement>('li.resource-item'))) {
      if (li.dataset.resid === resourceId) {
        item = li;
        break;
      }
    }
    const btn = item?.querySelector<HTMLButtonElement>('.knock-btn');
    if (btn) btn.disabled = true;
    showAlert('info', t('res.knocking', { id: resourceId }));
    try {
      const creds = await api.credentials();
      const handle = await createAgent(creds.privateKey, creds.nhp, creds.deviceId, 'error');
      const outcome = await knockResource(handle, creds.nhp, resourceId);
      handle.dispose();
      if (outcome.success && outcome.resourceHost) {
        showAlert('success', t('res.knockSuccess', { host: outcome.resourceHost }));
        const url = /^https?:\/\//.test(outcome.resourceHost)
          ? outcome.resourceHost
          : `https://${outcome.resourceHost}`;
        if (popup && !popup.closed) {
          // Break window.opener on the popup before navigating so the
          // AC page cannot reach back into this origin via tabnabbing.
          // Without this, a hostile AC could script a redirect on the
          // opener tab and harvest a fresh login state from the user
          // (the session cookie was just minted with a raw NHP private
          // key in flight, so the impact is not hypothetical).
          popup.opener = null;
          popup.location.href = url;
        } else {
          // Popup was blocked or already closed — fall back to navigating the
          // current tab so the knock isn't wasted.
          window.location.href = url;
        }
      } else {
        showAlert('error', t('res.knockFailed', { err: outcome.error ?? 'unknown' }));
        popup?.close();
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      showAlert('error', t('res.knockFailed', { err: msg }));
      popup?.close();
    } finally {
      if (btn) btn.disabled = false;
    }
  }
}
