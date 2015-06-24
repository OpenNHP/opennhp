'use strict';

var $ = require('jquery');
var UI = require('./core');

/**
 * @via https://github.com/Minwe/bootstrap/blob/master/js/alert.js
 * @copyright Copyright 2013 Twitter, Inc.
 * @license Apache 2.0
 */

// Alert Class
// NOTE: removeElement option is unavailable now
var Alert = function(element, options) {
  var _this = this;
  this.options = $.extend({}, Alert.DEFAULTS, options);
  this.$element = $(element);

  this.$element
    .addClass('am-fade am-in')
    .on('click.alert.amui', '.am-close', function() {
      _this.close();
    });
};

Alert.DEFAULTS = {
  removeElement: true
};

Alert.prototype.close = function() {
  var $element = this.$element;

  $element.trigger('close.alert.amui').removeClass('am-in');

  function processAlert() {
    $element.trigger('closed.alert.amui').remove();
  }

  UI.support.transition && $element.hasClass('am-fade') ?
    $element
      .one(UI.support.transition.end, processAlert)
      .emulateTransitionEnd(200) :
    processAlert();
};

// plugin
UI.plugin('alert', Alert);

// Init code
$(document).on('click.alert.amui.data-api', '[data-am-alert]', function(e) {
  var $target = $(e.target);
  $target.is('.am-close') && $(this).alert('close');
});

module.exports = Alert;
