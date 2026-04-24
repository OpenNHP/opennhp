// Theme toggle for docs.opennhp.org
// Flips between opennhp-dark and opennhp-light by swapping the href
// on the Just-the-Docs theme <link>. Persists choice in localStorage.

(function () {
  var STORAGE_KEY = 'opennhp-theme';
  var THEMES = { dark: 'opennhp-dark', light: 'opennhp-light' };
  var VALID_THEMES = ['opennhp-dark', 'opennhp-light'];

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
    /* Defense-in-depth: only accept the two known scheme names so an
       attacker-controlled value (e.g. via a prior XSS that manipulated
       the attribute) can't be written into a stylesheet href. */
    if (VALID_THEMES.indexOf(theme) === -1) return;
    document.documentElement.setAttribute('data-opennhp-theme', theme);
    try { localStorage.setItem(STORAGE_KEY, theme); } catch (e) { /* ignore quota/private-mode */ }

    var oldLink = stylesheetLink();
    if (!oldLink) return;
    var targetHref = '/assets/css/just-the-docs-' + theme + '.css';
    if (oldLink.getAttribute('href') === targetHref) return;

    /* Flicker-free swap: insert a parallel <link> and only remove the
       old one after the new sheet's load event. Both sheets are active
       during the overlap; the new one wins by source order. Matches
       the init pattern in head_custom.html. */
    var newLink = document.createElement('link');
    newLink.rel = 'stylesheet';
    newLink.href = targetHref;
    function cleanup() {
      if (oldLink && oldLink.parentNode) {
        oldLink.parentNode.removeChild(oldLink);
        oldLink = null;
      }
    }
    newLink.addEventListener('load', cleanup);
    newLink.addEventListener('error', cleanup);
    oldLink.parentNode.insertBefore(newLink, oldLink.nextSibling);
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
