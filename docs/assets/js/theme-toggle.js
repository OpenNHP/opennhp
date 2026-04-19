// Theme toggle for docs.opennhp.org
// Flips between opennhp-dark and opennhp-light by swapping the href
// on the Just-the-Docs theme <link>. Persists choice in localStorage.

(function () {
  var STORAGE_KEY = 'opennhp-theme';
  var THEMES = { dark: 'opennhp-dark', light: 'opennhp-light' };

  function currentTheme() {
    return document.documentElement.getAttribute('data-opennhp-theme') || THEMES.dark;
  }

  function stylesheetLink() {
    var links = document.getElementsByTagName('link');
    for (var i = 0; i < links.length; i++) {
      var href = links[i].getAttribute('href') || '';
      if (/\/just-the-docs-[a-z0-9-]+\.css$/.test(href) &&
          href.indexOf('just-the-docs-head-nav') === -1) {
        return links[i];
      }
    }
    return null;
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-opennhp-theme', theme);
    try { localStorage.setItem(STORAGE_KEY, theme); } catch (e) { /* ignore quota/private-mode */ }
    var link = stylesheetLink();
    if (link) {
      link.setAttribute('href', '/assets/css/just-the-docs-' + theme + '.css');
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
