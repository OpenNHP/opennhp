define(function(require, exports, module) {
  'use strict';

  require('core');
  require('ui.tabs');

  var $ = window.Zepto;

  function tabsInit() {
    $('[data-am-widget="tabs"]').each(function() {
      var options = $(this).data('amTabsNoswipe') ? {noSwipe: 1} : {};
      $(this).tabs(options);
    });
  }

  $(function() {
    tabsInit();
  });

  exports.init = tabsInit;
});
