'use strict';

var $ = require('jquery');
require('./core');
require('./ui.modal');
var addToHS = require('./ui.add2home');
var cookie = require('./util.cookie');

function footerInit() {
  // modal mode
  $('.am-footer-ysp').on('click', function() {
    $('#am-footer-modal').modal();
  });

  !window.AMUI_NO_ADD2HS && addToHS();

  // switch mode
  // switch to desktop
  $('[data-rel="desktop"]').on('click', function(e) {
    e.preventDefault();
    if (window.AMPlatform) { // front end
      window.AMPlatform.util.goDesktop();
    } else { // back end
      cookie.set('allmobilize', 'desktop', '', '/');
      window.location = window.location;
    }
  });
}

$(window).on('load', function() { // make sure `window.AMPlatform` is usable
  footerInit();
});

module.exports = $.AMUI.footer = {
  VERSION: '3.1.0',
  init: footerInit
};
