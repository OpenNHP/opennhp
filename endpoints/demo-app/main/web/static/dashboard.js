// dashboard.js — unlock the private key, list resources, knock on one.

import { NHPAgent } from '/static/nhp-agent.esm.js';
import { decryptPrivateKey } from '/static/crypto.js';

const { authServiceId, cipherScheme } = window.__NHP__ || {};

const unlockForm = document.getElementById('unlock-form');
const resourcesSection = document.getElementById('resources');
const resourceList = document.getElementById('resource-list');
const resourcesStatus = document.getElementById('resources-status');
const knockSection = document.getElementById('knock-result');
const knockHost = document.getElementById('knock-host');
const knockExpires = document.getElementById('knock-expires');
const errorEl = document.getElementById('dashboard-error');
const logoutBtn = document.getElementById('logout-btn');

unlockForm.addEventListener('submit', onUnlock);
logoutBtn.addEventListener('click', onLogout);

async function onUnlock(e) {
    e.preventDefault();
    errorEl.hidden = true;
    const fd = new FormData(e.target);
    const password = fd.get('password').toString();
    try {
        const [profile, config] = await Promise.all([
            fetchJson('/api/user/profile'),
            fetchJson('/api/nhp/config'),
        ]);

        if (!profile.encryptedPrivateKey || !profile.nhpPublicKey) {
            throw new Error('No NHP key registered for this account');
        }

        const privKey = await decryptPrivateKey(profile.encryptedPrivateKey, password);

        const agent = new NHPAgent({ privateKey: privKey, cipherScheme });
        await agent.init();
        agent.setIdentity({ userId: profile.userId, deviceId: 'browser-dash', organizationId: '' });
        agent.addServer({
            publicKey: config.serverPubKey,
            relayUrl: config.relayUrl,
            id: 'nhp-server',
        });

        unlockForm.hidden = true;
        await loadResources(agent);
    } catch (err) {
        console.error(err);
        errorEl.textContent = err.message || String(err);
        errorEl.hidden = false;
    }
}

async function loadResources(agent) {
    resourcesSection.hidden = false;
    resourcesStatus.textContent = 'Loading…';

    const list = await agent.listServices(authServiceId);
    if (!list.success) {
        resourcesStatus.textContent = `Failed to load resources: ${list.error || 'unknown error'}`;
        return;
    }
    const items = Object.entries(list.list || {});
    if (items.length === 0) {
        resourcesStatus.textContent = 'No resources available.';
        return;
    }
    resourcesStatus.hidden = true;
    resourceList.innerHTML = '';
    for (const [resId, info] of items) {
        const li = document.createElement('li');
        const name = document.createElement('span');
        name.textContent = info && info.hostname ? `${resId} (${info.hostname})` : resId;
        const btn = document.createElement('button');
        btn.className = 'btn primary';
        btn.textContent = 'Knock';
        btn.addEventListener('click', () => onKnock(agent, resId));
        li.appendChild(name);
        li.appendChild(btn);
        resourceList.appendChild(li);
    }
}

async function onKnock(agent, resourceId) {
    errorEl.hidden = true;
    knockSection.hidden = true;
    try {
        const result = await agent.knockResource({ resourceId, serviceId: authServiceId });
        if (!result.success) {
            throw new Error(result.error || 'Knock failed');
        }
        const hosts = Object.entries(result.resourceHosts || {})
            .map(([k, v]) => `${k}: ${v}`).join(', ') || '(no hosts returned)';
        knockHost.textContent = `Resources opened: ${hosts}`;
        knockExpires.textContent = new Date(result.expiresAt).toLocaleString();
        knockSection.hidden = false;
    } catch (err) {
        console.error(err);
        errorEl.textContent = err.message || String(err);
        errorEl.hidden = false;
    }
}

async function onLogout() {
    await fetch('/api/users/logout', { method: 'POST', credentials: 'include' });
    location.href = '/login';
}

async function fetchJson(url) {
    const res = await fetch(url, { credentials: 'include' });
    if (!res.ok) {
        const j = await res.json().catch(() => ({}));
        throw new Error(j.error || `request failed: ${res.status}`);
    }
    return res.json();
}
