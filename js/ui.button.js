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
  className: {
    loading: 'am-btn-loading',
    disabled: 'am-disabled'
  },
  spinner: undefined
};

Button.prototype.setState = function(state) {
  var disabled = 'disabled';
  var $element = this.$element;
  var options = this.options;
  var val = $element.is('input') ? 'val' : 'html';
  var loadingClassName = options.className.disabled + ' ' +
    options.className.loading;

  state = state + 'Text';

  if (!options.resetText) {
    options.resetText = $element[val]();
  }

  // add spinner for element with html()
  if (UI.support.animation && options.spinner &&
    val === 'html' && !this.hasSpinner) {
    options.loadingText = '<span class="am-icon-' +
    options.spinner +
    ' am-icon-spin"></span>' + options.loadingText;

    this.hasSpinner = true;
  }

  $element[val](options[state]);

  // push to event loop to allow forms to submit
  setTimeout($.proxy(function() {
    if (state == 'loadingText') {
      $element.addClass(loadingClassName).attr(disabled, disabled);
      this.isLoading = true;
    } else if (this.isLoading) {
      $element.removeClass(loadingClassName).removeAttr(disabled);
      this.isLoading = false;
    }
  }, this), 0);
};

Button.prototype.toggle = function() {
  var changed = true;
  var $element = this.$element;
  var $parent = this.$element.parent('.am-btn-group');

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

// Button plugin
function Plugin(option) {
  return this.each(function() {
    var $this = $(this);
    var data = $this.data('amui.button');
    var options = typeof option == 'object' && option || {};

    if (!data) {
      $this.data('amui.button', (data = new Button(this, options)));
    }

    if (option == 'toggle') {
      data.toggle();
    } else if (typeof option == 'string') {
      data.setState(option);
    }
  });
}

$.fn.button = Plugin;

// Init code
$(document).on('click.button.amui.data-api', '[data-am-button]', function(e) {
  var $btn = $(e.target);

  if (!$btn.hasClass('am-btn')) {
    $btn = $btn.closest('.am-btn');
  }

  Plugin.call($btn, 'toggle');
  e.preventDefault();
});

UI.ready(function(context) {
  $('[data-am-loading]', context).each(function() {
    $(this).button(UI.utils.parseOptions($(this).data('amLoading')));
  });
});

$.AMUI.button = Button;

module.exports = Button;
