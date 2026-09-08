// SPA entry. Determines whether the visitor has a session and shows
// resources, the login form, the registration form, or the
// complete-registration (resume) form accordingly.

import { api, ApiError } from './api.js';
import { getLang, setLang, t, type Lang } from './i18n.js';
import { renderLogin } from './views/login.js';
import { renderRegister } from './views/register.js';
import { renderResources } from './views/resources.js';
import { renderCompleteRegistration } from './views/complete-registration.js';

type View = 'loading' | 'login' | 'register' | 'resources' | 'complete-registration';

const root = document.querySelector<HTMLElement>('#app')!;

// Fixed language switcher — a globe-icon dropdown (EN ▾) mirroring the one
// on agent.opennhp.org. Picking a language persists the choice and reloads
// so every view re-renders in the chosen language from scratch.
function mountLangSwitcher(): void {
  if (document.getElementById('lang-switcher')) return;
  const langs: { code: Lang; short: string; name: string }[] = [
    { code: 'en', short: 'EN', name: 'English' },
    { code: 'zh-cn', short: '中文', name: '简体中文' },
  ];
  const cur = getLang();
  const shortLabel = langs.find((l) => l.code === cur)?.short ?? 'EN';

  const wrap = document.createElement('div');
  wrap.id = 'lang-switcher';
  wrap.className = 'lang-switcher';
  wrap.innerHTML = `
    <button id="langButton" type="button" aria-haspopup="menu" aria-expanded="false" aria-label="Language">
      <svg class="lang-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9"></circle><path d="M3 12h18"></path><path d="M12 3a14 14 0 0 1 0 18"></path><path d="M12 3a14 14 0 0 0 0 18"></path></svg>
      <span id="langButtonLabel">${shortLabel}</span><span class="caret">▾</span>
    </button>
    <ul id="langMenu" class="lang-menu" role="menu" hidden>
      ${langs.map((l) => `<li role="menuitem" data-lang="${l.code}" class="${l.code === cur ? 'active' : ''}">${l.name}</li>`).join('')}
    </ul>
  `;
  document.body.appendChild(wrap);

  const button = wrap.querySelector<HTMLButtonElement>('#langButton')!;
  const menu = wrap.querySelector<HTMLUListElement>('#langMenu')!;

  const closeMenu = (): void => {
    menu.hidden = true;
    button.setAttribute('aria-expanded', 'false');
  };
  button.addEventListener('click', (e) => {
    e.stopPropagation();
    const open = menu.hidden;
    menu.hidden = !open;
    button.setAttribute('aria-expanded', String(open));
  });
  menu.addEventListener('click', (e) => {
    const li = (e.target as HTMLElement).closest<HTMLLIElement>('li[data-lang]');
    if (!li) return;
    const lang = li.dataset.lang as Lang;
    if (lang === getLang()) {
      closeMenu();
      return;
    }
    setLang(lang);
    location.reload();
  });
  // Dismiss on outside click / Escape.
  document.addEventListener('click', closeMenu);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeMenu();
  });
}
mountLangSwitcher();
setLang(getLang()); // stamp <html lang> on first load
document.title = t('login.title');

async function detectSession(): Promise<{ username: string; email: string; status: string; cipherScheme: string; serverName: string; authProvider: string } | null> {
  try {
    return await api.me();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) return null;
    throw err;
  }
}

async function route(): Promise<void> {
  const next = await detectSession().then((me) => {
    if (!me) return 'login' as View;
    // A pending user finished account creation but never completed the
    // NHP_REG handshake — send them to the resume view, not resources.
    if (me.status === 'pending') return 'complete-registration' as View;
    return 'resources' as View;
  }, () => 'login' as View);
  show(next);
}

function show(view: View): void {
  switch (view) {
    case 'login':
      renderLogin(root, {
        onSignedIn: () => void route(),
        onSwitchToRegister: () => show('register'),
      });
      return;
    case 'register':
      renderRegister(root, {
        // Fresh registration sets a session on confirm; re-route so the
        // now-active user lands on resources instead of the login form.
        onRegistered: () => void route(),
        onSwitchToLogin: () => show('login'),
      });
      return;
    case 'complete-registration':
      void detectSession().then((me) => {
        if (!me) {
          show('login');
          return;
        }
        // Guard: if they're somehow active now, skip to resources.
        if (me.status !== 'pending') {
          show('resources');
          return;
        }
        renderCompleteRegistration(root, {
          username: me.username,
          email: me.email,
          onCompleted: () => void route(),
          onSignOut: () => show('login'),
        });
      });
      return;
    case 'resources':
      void detectSession().then((me) => {
        if (!me) {
          show('login');
          return;
        }
        if (me.status === 'pending') {
          show('complete-registration');
          return;
        }
        renderResources(root, {
          username: me.username,
          email: me.email,
          cipherScheme: me.cipherScheme,
          serverName: me.serverName,
          authProvider: me.authProvider,
          onSignOut: () => show('login'),
        });
      });
      return;
    case 'loading':
      root.innerHTML = `<div class="container"><p class="note">${t('common.loading')}</p></div>`;
      return;
  }
}

void route();
