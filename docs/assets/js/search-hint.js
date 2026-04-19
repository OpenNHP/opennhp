// Inject a platform-aware keyboard-shortcut hint (⌘ K / Ctrl K) into
// the search field's right side. External so jtd's HTML compressor can't
// mangle it.

(function () {
  function injectSearchHint() {
    var wrap = document.querySelector('.search-input-wrap');
    if (!wrap || wrap.querySelector('.search-kbd')) return;
    /* navigator.userAgent first — navigator.platform is deprecated and
       only kept as a fallback for older browsers without userAgent. */
    var mac = /Mac|iPhone|iPad|iPod/i.test(navigator.userAgent || navigator.platform || '');
    var kbd = document.createElement('span');
    kbd.className = 'search-kbd';
    kbd.setAttribute('aria-hidden', 'true');
    kbd.textContent = mac ? '⌘ K' : 'Ctrl K';
    wrap.appendChild(kbd);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', injectSearchHint);
  } else {
    injectSearchHint();
  }
})();
