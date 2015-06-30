'use strict';

var $ = require('jquery');
var UI = require('./core');

/**
 * @via https://github.com/twbs/bootstrap/blob/master/js/button.js
 * @copyright (c) 2011-2014 Twitter, Inc
 * @license The MIT License
 */

var Button = function(element, options) {
  this.$element = $(element);
  this.options = $.extend({}, Button.DEFAULTS, options);
  this.isLoading = false;
  this.hasSpinner = false;
};

Button.DEFAULTS = {
  loadingText: 'loading...',
  disabledClassName: 'am-disabled',
  spinner: undefined
};

Button.prototype.setState = function(state, stateText) {
  var $element = this.$element;
  var disabled = 'disabled';
  var data = $element.data();
  var options = this.options;
  var val = $element.is('input') ? 'val' : 'html';
  var stateClassName = 'am-btn-' + state + ' ' + options.disabledClassName;

  state += 'Text';

  if (!options.resetText) {
    options.resetText = $element[val]();
  }

  // add spinner for element with html()
  if (UI.support.animation && options.spinner &&
    val === 'html' && !this.hasSpinner) {
    options.loadingText = '<span class="am-icon-' + options.spinner +
      ' am-icon-spin"></span>' + options.loadingText;

    this.hasSpinner = true;
  }

  stateText = stateText ||
    (data[state] === undefined ? options[state] : data[state]);

  $element[val](stateText);

  // push to event loop to allow forms to submit
  setTimeout($.proxy(function() {
    // TODO: add stateClass for other states
    if (state === 'loadingText') {
      $element.addClass(stateClassName).attr(disabled, disabled);
      this.isLoading = true;
    } else if (this.isLoading) {
      $element.removeClass(stateClassName).removeAttr(disabled);
      this.isLoading = false;
    }
  }, this), 0);
};

Button.prototype.toggle = function() {
  var changed = true;
  var $element = this.$element;
  var $parent = this.$element.parent('[class*="am-btn-group"]');

  if ($parent.length) {
    var $input = this.$element.find('input');

    if ($input.prop('type') == 'radio') {
      if ($input.prop('checked') && $element.hasClass('am-active')) {
        changed = false;
      } else {
        $parent.find('.am-active').removeClass('am-active');
      }
    }

    if (changed) {
      $input.prop('checked',
        !$element.hasClass('am-active')).trigger('change');
    }
  }

  if (changed) {
    $element.toggleClass('am-active');
    if (!$element.hasClass('am-active')) {
      $element.blur();
    }
  }
};

UI.plugin('button', Button, {
  dataOptions: 'data-am-loading',
  methodCall: function(args, instance) {
    if (args[0] === 'toggle') {
      instance.toggle();
    } else if (typeof args[0] === 'string') {
      instance.setState.apply(instance, args);
    }
  }
});

// Init code
$(document).on('click.button.amui.data-api', '[data-am-button]', function(e) {
  e.preventDefault();
  var $btn = $(e.target);

  if (!$btn.hasClass('am-btn')) {
    $btn = $btn.closest('.am-btn');
  }

  $btn.button('toggle');
});

UI.ready(function(context) {
  $('[data-am-loading]', context).button();
});

module.exports = UI.button = Button;
