'use strict';

var $ = require('jquery');
require('./core');
require('./ui.pureview');
var IScroll = require('./ui.iscroll-lite');
var UI = $.AMUI;

/**
 * 表格滚动
 * @param index ID 标识，多个 paragraph 里面多个 table
 */
$.fn.scrollTable = function(index) {
  var $this = $(this);
  var $parent;

  $this.wrap('<div class="am-paragraph-table-container" ' +
  'id="am-paragraph-table-' + index + '">' +
  '<div class="am-paragraph-table-scroller"></div></div>');

  $parent = $this.parent();
  $parent.width($this.width());
  $parent.height($this.height());

  new IScroll('#am-paragraph-table-' + index, {
    eventPassthrough: true,
    scrollX: true,
    scrollY: false,
    preventDefault: false
  });
};

function paragraphInit() {
  var $paragraph = $('[data-am-widget="paragraph"]');

  $paragraph.each(function(index) {
    var $this = $(this);
    var options = UI.utils.parseOptions($this.attr('data-am-paragraph'));
    var $index = index;

    if (options.pureview) {
      $this.pureview();
    }

    if (options.tableScrollable) {
      $this.find('table').each(function(index) {
        if ($(this).width() > $(window).width()) {
          $(this).scrollTable($index + '-' + index);
        }
      });
    }
  });
}

$(window).on('load', function() {
  paragraphInit();
});

module.exports = $.AMUI.paragraph = {
  VERSION: '2.0.0',
  init: paragraphInit
};
