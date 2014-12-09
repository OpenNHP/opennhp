'use strict';

var $ = require('jquery');
require('./core');
require('./ui.smooth-scroll');

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

module.exports = $.AMUI.gotop = {
  VERSION: '4.0.1',
  init: goTopInit
};
