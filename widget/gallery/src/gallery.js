'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
require('../../../js/ui.pureview');

function galleryInit() {
  var $gallery = $('[data-am-widget="gallery"]');

  $gallery.each(function() {
    var options = UI.utils.parseOptions($(this).attr('data-am-gallery'));

    if (options.pureview) {
      (typeof options.pureview === 'object') ?
        $(this).pureview(options.pureview) : $(this).pureview();
    }
  });
}

$(galleryInit);

module.exports = UI.gallery = {
  VERSION: '3.0.0',
  init: galleryInit
};
