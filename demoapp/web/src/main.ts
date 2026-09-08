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

// Fixed language switcher (EN | 中文). Switching persists the choice and
// reloads so every view re-renders in the chosen language from scratch.
function mountLangSwitcher(): void {
  if (document.getElementById('lang-switcher')) return;
  const bar = document.createElement('div');
  bar.id = 'lang-switcher';
  const langs: Lang[] = ['en', 'zh-cn'];
  const labels: Record<Lang, string> = { en: 'English', 'zh-cn': '中文' };
  bar.innerHTML = langs
    .map((l) => `<button type="button" data-lang="${l}" class="${getLang() === l ? 'active' : ''}">${labels[l]}</button>`)
    .join('');
  bar.addEventListener('click', (e) => {
    const btn = (e.target as HTMLElement).closest<HTMLButtonElement>('button[data-lang]');
    if (!btn) return;
    const lang = btn.dataset.lang as Lang;
    if (lang === getLang()) return;
    setLang(lang);
    location.reload();
  });
  document.body.appendChild(bar);
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
