'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
var IScroll = require('../../../js/ui.iscroll-lite');
require('../../../js/ui.pureview');

/**
 * 表格滚动
 * @param {number} index ID 标识，多个 paragraph 里面多个 table
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

$(window).on('load', paragraphInit);

module.exports = UI.paragraph = {
  VERSION: '2.0.1',
  init: paragraphInit
};
