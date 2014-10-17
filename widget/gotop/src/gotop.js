define(function(require, exports, module) {
  'use strict';

  require('core');
  require('ui.smooth-scroll');

  var $ = window.Zepto;

  function goTopInit() {
    var $goTop = $('[data-am-widget="gotop"]');
    var $fixed = $goTop.filter('.am-gotop-fixed');
    var $win = $(window);

    $goTop.find('a').on('click', function(e) {
      e.preventDefault();
      $win.smoothScroll();
    });

    function checkPosition() {
      $fixed[($win.scrollTop() > 50 ? 'add' : 'remove') + 'Class']('am-active');
    }

    checkPosition();

    $win.on('scroll.gotop.amui', $.AMUI.utils.debounce(checkPosition, 100));
  }

  $(function() {
    goTopInit();
  });

  exports.init = goTopInit;
});
