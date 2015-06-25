'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
require('../../../js/ui.collapse');

function accordionInit() {
  var $accordion = $('[data-am-widget="accordion"]');
  var selector = {
    item: '.am-accordion-item',
    title: '.am-accordion-title',
    body: '.am-accordion-bd',
    disabled: '.am-disabled'
  };

  $accordion.each(function(i, item) {
    var options = UI.utils.parseOptions($(item).attr('data-am-accordion'));
    var $title = $(item).find(selector.title);

    $title.on('click.accordion.amui', function() {
      var $collapse = $(this).next(selector.body);
      var $parent = $(this).parent(selector.item);
      var data = $collapse.data('amui.collapse');

      if ($parent.is(selector.disabled)) {
        return;
      }

      $parent.toggleClass('am-active');

      if (!data) {
        $collapse.collapse();
      } else {
        $collapse.collapse('toggle');
      }

      !options.multiple &&
      $(item).children('.am-active').
        not($parent).not(selector.disabled).removeClass('am-active').
        find(selector.body + '.am-in').collapse('close');
    });
  });
}

// Init on DOM ready
$(accordionInit);

module.exports = UI.accordion = {
  VERSION: '2.1.0',
  init: accordionInit
};
