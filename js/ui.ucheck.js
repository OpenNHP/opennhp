'use strict';

var $ = require('jquery');
var UI = require('./core');

/**
 * UCheck
 * @via https://github.com/designmodo/Flat-UI/blob/8ef98df23ba7f5033e596a9bd05b53b535a9fe99/js/radiocheck.js
 * @license CC BY 3.0 & MIT
 * @param {HTMLElement} element
 * @param {object} options
 * @constructor
 */

var UCheck = function(element, options) {
  this.options = $.extend({}, UCheck.DEFAULTS, options);
  // this.options = $.extend({}, UCheck.DEFAULTS, this.$element.data(), options);
  this.$element = $(element);
  this.init();
};

UCheck.DEFAULTS = {
  checkboxClass: 'am-ucheck-checkbox',
  radioClass: 'am-ucheck-radio',
  checkboxTpl: '<span class="am-ucheck-icons">' +
  '<i class="am-icon-unchecked"></i><i class="am-icon-checked"></i></span>',
  radioTpl: '<span class="am-ucheck-icons">' +
  '<i class="am-icon-unchecked"></i><i class="am-icon-checked"></i></span>'
};

UCheck.prototype.init = function() {
  var $element = this.$element;
  var element = $element[0];
  var options = this.options;

  if (element.type === 'checkbox') {
    $element.addClass(options.checkboxClass)
      .after(options.checkboxTpl);
  } else if (element.type === 'radio') {
    $element.addClass(options.radioClass)
      .after(options.radioTpl);
  }
};

UCheck.prototype.check = function() {
  this.$element
    .prop('checked', true)
    .trigger('change.ucheck.amui')
    .trigger('checked.ucheck.amui');
},

UCheck.prototype.uncheck = function() {
  this.$element
    .prop('checked', false)
    // trigger `change` event for form validation, etc.
    // @see https://forum.jquery.com/topic/should-chk-prop-checked-true-trigger-change-event
    .trigger('change')
    .trigger('unchecked.ucheck.amui');
},

UCheck.prototype.toggle = function() {
  this.$element.
    prop('checked', function(i, value) {
      return !value;
    })
    .trigger('change.ucheck.amui')
    .trigger('toggled.ucheck.amui');
},

UCheck.prototype.disable = function() {
  this.$element
    .prop('disabled', true)
    .trigger('change.ucheck.amui')
    .trigger('disabled.ucheck.amui');
},

UCheck.prototype.enable = function() {
  this.$element.prop('disabled', false);
  this.$element.trigger('change.ucheck.amui').trigger('enabled.ucheck.amui');
},

UCheck.prototype.destroy = function() {
  this.$element
    .removeData('amui.ucheck')
    .removeClass(this.options.checkboxClass + ' ' + this.options.radioClass)
    .next('.am-ucheck-icons')
    .remove()
  .end()
    .trigger('destroyed.ucheck.amui');
};

UI.plugin('uCheck', UCheck, {
  after: function() {
    // Adding 'am-nohover' class for touch devices
    if (UI.support.touch) {
      this.parent().hover(function() {
        $(this).addClass('am-nohover');
      }, function() {
        $(this).removeClass('am-nohover');
      });
    }
  }
});

UI.ready(function(context) {
  $('[data-am-ucheck]', context).uCheck();
});

module.exports = UCheck;

// TODO: 与表单验证结合使用的情况
