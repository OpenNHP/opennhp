// Theme toggle for docs.opennhp.org
// Flips between opennhp-dark and opennhp-light, persists to localStorage,
// and uses jtd.setTheme() (the Just-the-Docs runtime API) to swap the
// active color scheme stylesheet without reloading.

(function () {
  var STORAGE_KEY = 'opennhp-theme';
  var THEMES = { dark: 'opennhp-dark', light: 'opennhp-light' };

  function currentTheme() {
    return document.documentElement.getAttribute('data-opennhp-theme') || THEMES.dark;
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-opennhp-theme', theme);
    try { localStorage.setItem(STORAGE_KEY, theme); } catch (e) { /* ignore quota/private-mode */ }
    if (window.jtd && typeof window.jtd.setTheme === 'function') {
      window.jtd.setTheme(theme);
    }
  }

  function init() {
    var btn = document.getElementById('opennhp-theme-toggle');
    if (!btn) return;
    btn.addEventListener('click', function () {
      var next = currentTheme() === THEMES.dark ? THEMES.light : THEMES.dark;
      applyTheme(next);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
