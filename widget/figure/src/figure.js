'use strict';

require('./core');
require('./ui.pureview');
var $ = require('jquery');
var UI = $.AMUI;

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

    if (options.pureview) {
      if (options.pureview === 'auto') {
        var zoomAble = $.isImgZoomAble($(item).find('img')[0]);
        zoomAble && $(item).pureview();
      } else {
        $(item).addClass('am-figure-zoomable').pureview();
      }
    }
  });
}

$(window).on('load', function() {
  figureInit();
});

module.exports = $.AMUI.figure = {
  VERSION: '2.0.1',
  init: figureInit
};
