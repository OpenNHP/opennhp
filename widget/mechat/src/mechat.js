'use strict';

var $ = require('jquery');
require('./core');

function mechatInit() {
  if (!$('#mechat').length) {
    return;
  }

  var $mechat = $('[data-am-widget="mechat"]');
  var unitid = $mechat.data('am-mechat-unitid');
  var $mechatData = $('<script>', {
    charset: 'utf-8',
    src: 'http://mechatim.com/js/unit/button.js?id=' + unitid
  });

  $('body').append($mechatData);
}

// Lazy load
$(window).on('load', mechatInit);

module.exports = $.AMUI.mechat = {
  VERSION: '2.0.0',
  init: mechatInit
};
