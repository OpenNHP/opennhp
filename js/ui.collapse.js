'use strict';

var $ = require('jquery');
var UI = require('./core');

/**
 * @via https://github.com/twbs/bootstrap/blob/master/js/collapse.js
 * @copyright (c) 2011-2014 Twitter, Inc
 * @license The MIT License
 */

var Collapse = function(element, options) {
  this.$element = $(element);
  this.options = $.extend({}, Collapse.DEFAULTS, options);
  this.transitioning = null;

  if (this.options.parent) {
    this.$parent = $(this.options.parent);
  }

  if (this.options.toggle) {
    this.toggle();
  }
};

Collapse.DEFAULTS = {
  toggle: true
};

Collapse.prototype.open = function() {
  if (this.transitioning || this.$element.hasClass('am-in')) {
    return;
  }

  var startEvent = $.Event('open.collapse.amui');
  this.$element.trigger(startEvent);

  if (startEvent.isDefaultPrevented()) {
    return;
  }

  var actives = this.$parent && this.$parent.find('> .am-panel > .am-in');

  if (actives && actives.length) {
    var hasData = actives.data('amui.collapse');

    if (hasData && hasData.transitioning) {
      return;
    }

    Plugin.call(actives, 'close');

    hasData || actives.data('amui.collapse', null);
  }

  this.$element
    .removeClass('am-collapse')
    .addClass('am-collapsing').height(0);

  this.transitioning = 1;

  var complete = function() {
    this.$element
      .removeClass('am-collapsing')
      .addClass('am-collapse am-in')
      .height('')
      .trigger('opened.collapse.amui');
    this.transitioning = 0;
  };

  if (!UI.support.transition) {
    return complete.call(this);
  }

  var scrollHeight = this.$element[0].scrollHeight;

  this.$element
    .one(UI.support.transition.end, $.proxy(complete, this))
    .emulateTransitionEnd(300)
    .css({height: scrollHeight}); // 当折叠的容器有 padding 时，如果用 height() 只能设置内容的宽度
};

Collapse.prototype.close = function() {
  if (this.transitioning || !this.$element.hasClass('am-in')) {
    return;
  }

  var startEvent = $.Event('close.collapse.amui');
  this.$element.trigger(startEvent);

  if (startEvent.isDefaultPrevented()) {
    return;
  }

  this.$element.height(this.$element.height()).redraw();

  this.$element.addClass('am-collapsing').
    removeClass('am-collapse am-in');

  this.transitioning = 1;

  var complete = function() {
    this.transitioning = 0;
    this.$element
      .trigger('closed.collapse.amui')
      .removeClass('am-collapsing')
      .addClass('am-collapse');
    // css({height: '0'});
  };

  if (!UI.support.transition) {
    return complete.call(this);
  }

  this.$element.height(0)
    .one(UI.support.transition.end, $.proxy(complete, this))
    .emulateTransitionEnd(300);
};

Collapse.prototype.toggle = function() {
  this[this.$element.hasClass('am-in') ? 'close' : 'open']();
};

// Collapse Plugin
function Plugin(option) {
  return this.each(function() {
    var $this = $(this);
    var data = $this.data('amui.collapse');
    var options = $.extend({}, Collapse.DEFAULTS,
      UI.utils.options($this.attr('data-am-collapse')),
      typeof option == 'object' && option);

    if (!data && options.toggle && option === 'open') {
      option = !option;
    }

    if (!data) {
      $this.data('amui.collapse', (data = new Collapse(this, options)));
    }

    if (typeof option == 'string') {
      data[option]();
    }
  });
}

$.fn.collapse = Plugin;

// Init code
$(document).on('click.collapse.amui.data-api', '[data-am-collapse]',
  function(e) {
    var href;
    var $this = $(this);
    var options = UI.utils.options($this.attr('data-am-collapse'));
    var target = options.target ||
      e.preventDefault() ||
      (href = $this.attr('href')) &&
      href.replace(/.*(?=#[^\s]+$)/, '');
    var $target = $(target);
    var data = $target.data('amui.collapse');
    var option = data ? 'toggle' : options;
    var parent = options.parent;
    var $parent = parent && $(parent);

    if (!data || !data.transitioning) {
      if ($parent) {
        // '[data-am-collapse*="{parent: \'' + parent + '"]
        $parent.find('[data-am-collapse]').not($this).addClass('am-collapsed');
      }

      $this[$target.hasClass('am-in') ?
        'addClass' : 'removeClass']('am-collapsed');
    }

    Plugin.call($target, option);
  });

module.exports = UI.collapse = Collapse;

// TODO: 更好的 target 选择方式
//       折叠的容器必须没有 border/padding 才能正常处理，否则动画会有一些小问题
//       寻找更好的未知高度 transition 动画解决方案，max-height 之类的就算了
