'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
require('../../../js/ui.tabs');

function tabsInit() {
  $('[data-am-widget="tabs"]').each(function() {
    var options = $(this).data('amTabsNoswipe') ? {noSwipe: 1} : {};
    $(this).tabs(options);
  });
}

$(tabsInit);

module.exports = UI.tab = {
  VERSION: '4.0.1',
  init: tabsInit
};
