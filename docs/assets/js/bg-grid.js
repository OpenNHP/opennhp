// Atmospheric background grid. Styling in custom.scss scopes visibility
// to the dark scheme, so light-mode users see nothing. External file to
// avoid the HTML compressor mangling multi-line inline scripts.

(function () {
  function injectLayers() {
    if (document.querySelector('.opennhp-bg-grid')) return;
    var grid = document.createElement('div');
    grid.className = 'opennhp-bg-grid';
    document.body.insertBefore(grid, document.body.firstChild);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', injectLayers);
  } else {
    injectLayers();
  }
})();
