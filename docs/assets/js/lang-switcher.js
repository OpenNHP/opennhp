// Language switcher for docs.opennhp.org
// Dropdown menu with English / 简体中文. Clicking an option navigates to
// the equivalent page in the other language when possible, or falls back
// to that language's root.

(function () {
  var ZH_PREFIX = '/zh-cn';
  // There is no /zh-cn/ landing page — the Chinese "Overview" equivalent
  // lives at /zh-cn/overview/. Map the English root ↔ Chinese overview.
  var ZH_OVERVIEW = '/zh-cn/overview/';

  /* Pages that exist in only one language. Selecting the other
     language from one of these sends the user to that language's
     root/overview instead of a non-existent path. Keep these lists
     in sync with the docs/ filesystem. */
  var ENGLISH_ONLY_PATHS = [];
  var CHINESE_ONLY_PATHS = [
    '/zh-cn/claw-dhp-demo/'
  ];

  function targetUrl(lang) {
    var path = window.location.pathname;
    var onZh = path.indexOf(ZH_PREFIX + '/') === 0;

    if (lang === 'zh-cn') {
      if (onZh) return path;
      if (path === '/' || path === '') return ZH_OVERVIEW;
      if (ENGLISH_ONLY_PATHS.indexOf(path) !== -1) return ZH_OVERVIEW;
      return ZH_PREFIX + path;
    }
    // lang === 'en'
    if (!onZh) return path;
    if (path === ZH_OVERVIEW) return '/';
    if (CHINESE_ONLY_PATHS.indexOf(path) !== -1) return '/';
    var stripped = path.slice(ZH_PREFIX.length) || '/';
    return stripped;
  }

  function closeMenu(button, menu) {
    menu.hidden = true;
    button.setAttribute('aria-expanded', 'false');
  }

  function openMenu(button, menu) {
    menu.hidden = false;
    button.setAttribute('aria-expanded', 'true');
  }

  function init() {
    var button = document.getElementById('opennhp-lang-button');
    var menu = document.getElementById('opennhp-lang-menu');
    if (!button || !menu) return;

    button.addEventListener('click', function (e) {
      e.stopPropagation();
      if (menu.hidden) openMenu(button, menu); else closeMenu(button, menu);
    });

    menu.addEventListener('click', function (e) {
      var item = e.target.closest('[data-lang]');
      if (!item) return;
      window.location.href = targetUrl(item.getAttribute('data-lang'));
    });

    // Close on outside click
    document.addEventListener('click', function (e) {
      if (!menu.hidden && !menu.contains(e.target) && e.target !== button) {
        closeMenu(button, menu);
      }
    });

    // Close on Escape
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && !menu.hidden) {
        closeMenu(button, menu);
        button.focus();
      }
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
