// Minimal i18n for the demo app: English + Simplified Chinese.
//
// Strings are looked up by key via t(key, vars?). Values may contain
// {placeholder} tokens replaced from `vars`. A handful of footer strings
// carry inline HTML (links) and are injected via innerHTML by the caller;
// everything else is plain text. The active language is persisted in
// localStorage and defaults to the browser's preferred language.

export type Lang = 'en' | 'zh-cn';

type Dict = Record<string, string>;

const EN: Dict = {
  // Shared
  'common.signedInAs': 'Signed in as',
  'common.signOut': 'Sign out',
  'common.loading': 'Loading…',
  'common.subtitle': 'A working example of adding the OpenNHP to an existing web application',

  // Language switcher
  'lang.label': 'Language',
  'lang.en': 'English',
  'lang.zh': '中文',

  // Login
  'login.title': 'OpenNHP Login Integration Demo App',
  'login.credTitle': 'User credentials',
  'login.credHint': 'Sign in with a username and password, or create a new account.',
  'login.username': 'Username',
  'login.password': 'Password',
  'login.signIn': 'Sign in',
  'login.createAccount': 'New here? Create an account',
  'login.oauthTitle': 'Single Sign-On',
  'login.oauthHint': 'Social Login',
  'login.github': 'Sign in with GitHub',
  'login.errRequired': 'Username and password are required.',
  'login.signingIn': 'Signing in…',
  'login.failed': 'Sign in failed: {msg}',

  // Architecture diagram
  'arch.summary': 'How the OpenNHP Integration Demo works',

  // Footer
  'footer.detectingIp': 'Detecting IP…',
  'footer.ipUnavailable': 'IP detection unavailable',
  'footer.poweredBy': 'Powered by <a href="https://github.com/OpenNHP/opennhp" target="_blank" rel="noopener"><strong>OpenNHP</strong></a> &mdash; Network-infrastructure Hiding Protocol',
  'footer.sponsoredBy': 'Sponsored by: <a href="https://layerv.ai" target="_blank" rel="noopener">LayerV.ai</a>',
  'footer.viewSource': "View this demo app's source on GitHub",

  // Register
  'reg.accountCreds': '1. Account credentials',
  'reg.username': 'Username',
  'reg.email': 'Email',
  'reg.password': 'Password (min 8 chars)',
  'reg.nhpServer': 'NHP server',
  'reg.cipherScheme': 'Cipher scheme',
  'reg.submit': 'Create account & request OTP',
  'reg.switchLogin': 'Already have an account? Sign in',
  'reg.noServers': 'No servers configured',
  'reg.allRequired': 'All fields are required.',
  'reg.selectServerScheme': 'Select an NHP server and cipher scheme.',
  'reg.creating': 'Creating account and generating NHP key pair…',
  'reg.failed': 'Registration failed: {msg}',
  'reg.loadingScheme': 'Loading…',

  // NHP registration panel
  'nhp.title': 'NHP registration',
  'nhp.back': 'Back',
  'nhp.otpSentTo': 'An OTP has been sent to <strong>{email}</strong>. In docker environments, check the nhp-server logs for the OTP fallback.',
  'nhp.otpLabel': 'OTP',
  'nhp.registerBtn': 'Register public key with nhp-server',
  'nhp.resend': 'Resend code',
  'nhp.resendCooldown': 'Resend code ({n}s)',
  'nhp.agentNotReady': 'Agent not ready. Go back and try again.',
  'nhp.requestingNew': 'Requesting a new OTP…',
  'nhp.newOtpSent': 'A new OTP has been sent to {email}.',
  'nhp.otpReqFailed': 'NHP-OTP request failed{suffix}.',
  'nhp.enterOtp': 'Enter the OTP from your email.',
  'nhp.driving': 'Driving NHP-REG handshake…',
  'nhp.regFailedCheck': 'NHP registration failed — check the OTP and try again.',
  'nhp.complete': 'Registration complete.',
  'nhp.handshakeFailed': 'NHP handshake failed: {msg}',
  'nhp.creatingAgent': 'Creating NHP agent and requesting OTP…',

  // Complete registration
  'cr.title': 'Complete NHP Registration',
  'cr.subtitle': 'Your account exists but the NHP key was never registered with nhp-server. Pick your cluster and cipher scheme, then complete the handshake to activate your account.',
  'cr.binding': 'NHP binding',
  'cr.cluster': 'NHP server cluster',
  'cr.cipherScheme': 'Cipher scheme',
  'cr.confirm': 'Confirm & request OTP',
  'cr.noClusters': 'No nhp-server clusters are configured.',
  'cr.selectClusterScheme': 'Select an NHP server cluster and cipher scheme.',
  'cr.generating': 'Generating NHP key material under the selected binding…',
  'cr.clickRetry': 'Click "Confirm & request OTP" to retry.',
  'cr.bindingFailed': 'Binding failed: {msg}',
  'cr.loadCatalogFailed': 'Failed to load server catalog: {msg}',

  // Resources
  'res.title': 'Protected Resources',
  'res.subtitle': 'Click "Access" to knock nhp-server. The protected resource is hidden until the knock opens the door.',
  'res.deleteAccount': 'Delete account',
  'res.badgeAuth': 'Auth',
  'res.badgeAlg': 'Alg',
  'res.badgeServer': 'Server',
  'res.access': 'Access',
  'res.providerLocal': 'Local',
  'res.deleteConfirm': 'Delete your account? This permanently removes your credentials and NHP key material from this demo. You can re-register afterward. This cannot be undone.',
  'res.fetchingCreds': 'Fetching credentials from server…',
  'res.loadFailed': 'Failed to load: {msg}',
  'res.initAgent': 'Initializing NHP agent and listing services…',
  'res.none': 'No accessible resources. Confirm registration completed and that the nhp-server basic plugin allows your user.',
  'res.knocking': 'Knocking {id}…',
  'res.knockSuccess': 'Knock successful — opening {host}',
  'res.knockFailed': 'Knock failed: {err}',
  'res.listFailed': 'listServices failed: {msg}',
  'res.deleteFailed': 'Failed to delete account: {msg}',
  'res.idPrefix': 'id',
};

const ZH: Dict = {
  'common.signedInAs': '已登录：',
  'common.signOut': '退出登录',
  'common.loading': '加载中…',
  'common.subtitle': '将 OpenNHP 集成到现有 Web 应用的可运行示例',

  'lang.label': '语言',
  'lang.en': 'English',
  'lang.zh': '中文',

  'login.title': 'OpenNHP 登录集成演示应用',
  'login.credTitle': '用户凭据',
  'login.credHint': '使用用户名和密码登录，或创建新账户。',
  'login.username': '用户名',
  'login.password': '密码',
  'login.signIn': '登录',
  'login.createAccount': '新用户？创建账户',
  'login.oauthTitle': 'OAuth / SAML',
  'login.oauthHint': '使用联合身份提供商登录。',
  'login.github': '使用 GitHub 登录',
  'login.errRequired': '用户名和密码为必填项。',
  'login.signingIn': '正在登录…',
  'login.failed': '登录失败：{msg}',

  'arch.summary': 'OpenNHP 集成演示的工作原理',

  'footer.detectingIp': '正在检测 IP…',
  'footer.ipUnavailable': '无法检测 IP',
  'footer.poweredBy': '驱动技术：<a href="https://github.com/OpenNHP/opennhp" target="_blank" rel="noopener"><strong>OpenNHP</strong></a> &mdash; 网络基础设施隐藏协议',
  'footer.sponsoredBy': '赞助方：<a href="https://layerv.ai" target="_blank" rel="noopener">LayerV.ai</a>',
  'footer.viewSource': '在 GitHub 上查看此演示应用的源代码',

  'reg.accountCreds': '1. 账户凭据',
  'reg.username': '用户名',
  'reg.email': '电子邮箱',
  'reg.password': '密码（至少 8 个字符）',
  'reg.nhpServer': 'NHP 服务器',
  'reg.cipherScheme': '加密方案',
  'reg.submit': '创建账户并请求 OTP',
  'reg.switchLogin': '已有账户？登录',
  'reg.noServers': '未配置服务器',
  'reg.allRequired': '所有字段均为必填项。',
  'reg.selectServerScheme': '请选择 NHP 服务器和加密方案。',
  'reg.creating': '正在创建账户并生成 NHP 密钥对…',
  'reg.failed': '注册失败：{msg}',
  'reg.loadingScheme': '加载中…',

  'nhp.title': 'NHP 注册',
  'nhp.back': '返回',
  'nhp.otpSentTo': '一次性验证码（OTP）已发送至 <strong>{email}</strong>。在 docker 环境中，可在 nhp-server 日志中查看 OTP 回退值。',
  'nhp.otpLabel': 'OTP',
  'nhp.registerBtn': '向 nhp-server 注册公钥',
  'nhp.resend': '重新发送验证码',
  'nhp.resendCooldown': '重新发送验证码（{n}秒）',
  'nhp.agentNotReady': '代理未就绪。请返回后重试。',
  'nhp.requestingNew': '正在请求新的 OTP…',
  'nhp.newOtpSent': '新的 OTP 已发送至 {email}。',
  'nhp.otpReqFailed': 'NHP-OTP 请求失败{suffix}。',
  'nhp.enterOtp': '请输入邮件中的 OTP。',
  'nhp.driving': '正在执行 NHP-REG 握手…',
  'nhp.regFailedCheck': 'NHP 注册失败 —— 请检查 OTP 后重试。',
  'nhp.complete': '注册完成。',
  'nhp.handshakeFailed': 'NHP 握手失败：{msg}',
  'nhp.creatingAgent': '正在创建 NHP 代理并请求 OTP…',

  'cr.title': '完成 NHP 注册',
  'cr.subtitle': '您的账户已存在，但 NHP 密钥尚未向 nhp-server 注册。请选择集群和加密方案，然后完成握手以激活账户。',
  'cr.binding': 'NHP 绑定',
  'cr.cluster': 'NHP 服务器集群',
  'cr.cipherScheme': '加密方案',
  'cr.confirm': '确认并请求 OTP',
  'cr.noClusters': '未配置 nhp-server 集群。',
  'cr.selectClusterScheme': '请选择 NHP 服务器集群和加密方案。',
  'cr.generating': '正在根据所选绑定生成 NHP 密钥材料…',
  'cr.clickRetry': '点击“确认并请求 OTP”重试。',
  'cr.bindingFailed': '绑定失败：{msg}',
  'cr.loadCatalogFailed': '加载服务器目录失败：{msg}',

  'res.title': '受保护的资源',
  'res.subtitle': '点击“访问”以叩门 nhp-server。在叩门开门之前，受保护的资源处于隐藏状态。',
  'res.deleteAccount': '删除账户',
  'res.badgeAuth': '认证',
  'res.badgeAlg': '算法',
  'res.badgeServer': '服务器',
  'res.access': '访问',
  'res.providerLocal': '本地',
  'res.deleteConfirm': '确定要删除您的账户吗？此操作将从本演示中永久移除您的凭据和 NHP 密钥材料。之后您可以重新注册。此操作无法撤销。',
  'res.fetchingCreds': '正在从服务器获取凭据…',
  'res.loadFailed': '加载失败：{msg}',
  'res.initAgent': '正在初始化 NHP 代理并列出服务…',
  'res.none': '没有可访问的资源。请确认注册已完成，且 nhp-server basic 插件已允许您的用户。',
  'res.knocking': '正在叩门 {id}…',
  'res.knockSuccess': '叩门成功 —— 正在打开 {host}',
  'res.knockFailed': '叩门失败：{err}',
  'res.listFailed': 'listServices 失败：{msg}',
  'res.deleteFailed': '删除账户失败：{msg}',
  'res.idPrefix': 'id',
};

const DICTS: Record<Lang, Dict> = { en: EN, 'zh-cn': ZH };

const STORAGE_KEY = 'demoapp.lang';

function detectLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'en' || saved === 'zh-cn') return saved;
  } catch { /* ignore */ }
  const nav = (navigator.language || '').toLowerCase();
  return nav.startsWith('zh') ? 'zh-cn' : 'en';
}

let current: Lang = detectLang();

export function getLang(): Lang {
  return current;
}

export function setLang(lang: Lang): void {
  current = lang;
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch { /* ignore */ }
  try {
    document.documentElement.lang = lang === 'zh-cn' ? 'zh-CN' : 'en';
  } catch { /* ignore */ }
}

// Translate `key`, interpolating {name} tokens from `vars`. Falls back to
// the English string, then to the raw key, so a missing translation is
// visible but never blank.
export function t(key: string, vars?: Record<string, string | number>): string {
  const raw = DICTS[current][key] ?? EN[key] ?? key;
  if (!vars) return raw;
  return raw.replace(/\{(\w+)\}/g, (_, name) =>
    name in vars ? String(vars[name]) : `{${name}}`,
  );
}

// Render the globe-icon language dropdown (EN ▾ / 中文 ▾) into `container`
// (a view's .container). Mirrors the switcher on agent.opennhp.org.
// Picking a language persists it and reloads so every view re-renders in
// the chosen language.
export function renderLangSwitcher(container: HTMLElement): void {
  const langs: { code: Lang; short: string; name: string }[] = [
    { code: 'en', short: 'EN', name: 'English' },
    { code: 'zh-cn', short: '中文', name: '简体中文' },
  ];
  const cur = getLang();
  const shortLabel = langs.find((l) => l.code === cur)?.short ?? 'EN';

  const wrap = document.createElement('div');
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
  // Prepend so it anchors to the container's top-right regardless of view.
  container.insertBefore(wrap, container.firstChild);

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
  document.addEventListener('click', closeMenu);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeMenu();
  });
}
