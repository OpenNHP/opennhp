define(function(require, exports, module) {
  'use strict';

  require('core');

  var $ = window.Zepto,
      UI = $.AMUI,
      IScroll = require('ui.iscroll-lite'),
      paragraphInit;

  /**
   * 表格滚动
   * @param index ID 标识，多个 paragraph 里面多个 table
   */
  $.fn.scrollTable = function(index) {
    var $this = $(this),
        $parent;

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

  paragraphInit = function() {
    var $paragraph = $('[data-am-widget="paragraph"]');

    $paragraph.each(function(index) {
      var $this = $(this),
          options = UI.utils.parseOptions($this.attr('data-am-paragraph')),
          $index = index;

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

  };

  $(window).on('load', function() {
    paragraphInit();
  });

  exports.init = paragraphInit;
});
