'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
require('ui.pureview');

/**
 * Is Images zoomable
 * @return {Boolean}
 */
$.isImgZoomAble = function(element) {
  var t = new Image();
  t.src = element.src;

  var zoomAble = ($(element).width() < t.width);

  if (zoomAble) {
    $(element).closest('.am-figure').addClass('am-figure-zoomable');
  }

  return zoomAble;
};

function figureInit() {
  $('.am-figure').each(function(i, item) {
    var options = UI.utils.parseOptions($(item).attr('data-am-figure'));
    var $item = $(item);
    var data;

    if (options.pureview) {
      if (options.pureview === 'auto') {
        var zoomAble = $.isImgZoomAble($item.find('img')[0]);
        zoomAble && $item.pureview();
      } else {
        $item.addClass('am-figure-zoomable').pureview();
      }
    }

    data = $item.data('amui.pureview');

    if (data) {
      $item.on('click', ':not(img)', function() {
        data.open(0);
      });
    }
  });
}

$(window).on('load', figureInit);

module.exports = UI.figure = {
  VERSION: '2.0.3',
  init: figureInit
};
