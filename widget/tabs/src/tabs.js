'use strict';

var $ = require('jquery');
require('./core');
require('./ui.tabs');

function tabsInit() {
  $('[data-am-widget="tabs"]').each(function() {
    var options = $(this).data('amTabsNoswipe') ? {noSwipe: 1} : {};
    $(this).tabs(options);
  });
}

$(function() {
  tabsInit();
});

module.exports = $.AMUI.tab = {
  VERSION: '4.0.0',
  init: tabsInit
};
