'use strict';

var $ = require('jquery');
require('../../../js/core');

function headerInit() {
  $('[data-am-widget="header"]').each(function() {
    if ($(this).hasClass('am-header-fixed')) {
      $('body').addClass('am-with-fixed-header');
      return false;
    }
  });
}

$(headerInit);

module.exports = $.AMUI.header = {
  VERSION: '2.0.0',
  init: headerInit
};
