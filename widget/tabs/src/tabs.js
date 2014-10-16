define(function(require, exports, module) {
  'use strict';

  require('core');
  require('ui.tabs');

  var $ = window.Zepto;

  function tabsInit() {
    $('[data-am-widget="tabs"]').tabs();
  }

  $(function() {
    tabsInit();
  });

  exports.init = tabsInit;
});
