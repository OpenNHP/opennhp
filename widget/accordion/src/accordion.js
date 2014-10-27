define(function(require, exports, module) {
  'use strict';

  require('core');
  require('ui.collapse');

  var $ = window.Zepto;
  var UI = $.AMUI;

  function accordionInit() {
    var $accordion = $('[data-am-widget="accordion"]');
    var selector = {
      item: '.am-accordion-item',
      title: '.am-accordion-title',
      content: '.am-accordion-content'
    };

    $accordion.each(function(i, item) {
      var options = UI.utils.parseOptions($(item).attr('data-am-accordion'));
      var $title = $(item).find(selector.title);

      $title.on('click.accordion.amui', function() {
        var $content = $(this).next(selector.content);
        var $parent = $(this).parent(selector.item);
        var data = $content.data('amui.collapse');

        $parent.toggleClass('am-active');

        if (!data) {
          $content.collapse();
        } else {
          $content.collapse('toggle');
        }

        !options.multiple &&
        $(item).children('.am-active').
            not($parent).removeClass('am-active').
            find('.am-accordion-content.am-in').collapse('close');
      });
    });
  }

  // Init on DOM ready
  $(function() {
    accordionInit();
  });

  exports.init = accordionInit;
});
