'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');
require('../../../js/ui.flexslider');

function sliderInit() {
  var $sliders = $('[data-am-widget="slider"]');
  $sliders.not('.am-slider-manual').each(function(i, item) {
    var options = UI.utils.parseOptions($(item).attr('data-am-slider'));
    $(item).flexslider(options);
  });
}

$(sliderInit);

module.exports = UI.slider = {
  VERSION: '3.0.1',
  init: sliderInit
};
