'use strict';

var $ = require('jquery');
var UI = require('./core');
require('./util.hammer');

var $win = $(window);
var $doc = $(document);
var scrollPos;

/**
 * @via https://github.com/uikit/uikit/blob/master/src/js/offcanvas.js
 * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
 */

var OffCanvas = function(element, options) {
  this.$element = $(element);
  this.options = $.extend({}, OffCanvas.DEFAULTS, options);
  this.active = null;
  this.bindEvents();
};

OffCanvas.DEFAULTS = {
  duration: 300,
  effect: 'overlay' // {push|overlay}, push is too expensive
};

OffCanvas.prototype.open = function(relatedElement) {
  var _this = this;
  var $element = this.$element;

  if (!$element.length || $element.hasClass('am-active')) {
    return;
  }

  var effect = this.options.effect;
  var $html = $('html');
  var $body = $('body');
  var $bar = $element.find('.am-offcanvas-bar').first();
  var dir = $bar.hasClass('am-offcanvas-bar-flip') ? -1 : 1;

  $bar.addClass('am-offcanvas-bar-' + effect);

  scrollPos = {x: window.scrollX, y: window.scrollY};

  $element.addClass('am-active');

  $body.css({
    width: window.innerWidth,
    height: $win.height()
  }).addClass('am-offcanvas-page');

  if (effect !== 'overlay') {
    $body.css({
      'margin-left': $bar.outerWidth() * dir
    }).width(); // force redraw
  }

  $html.css('margin-top', scrollPos.y * -1);

  setTimeout(function() {
    $bar.addClass('am-offcanvas-bar-active').width();
  }, 0);

  $element.trigger('open.offcanvas.amui');

  this.active = 1;

  // Close OffCanvas when none content area clicked
  $element.on('click.offcanvas.amui', function(e) {
    var $target = $(e.target);

    if ($target.hasClass('am-offcanvas-bar')) {
      return;
    }

    if ($target.parents('.am-offcanvas-bar').first().length) {
      return;
    }

    // https://developer.mozilla.org/zh-CN/docs/DOM/event.stopImmediatePropagation
    e.stopImmediatePropagation();

    _this.close();
  });

  $html.on('keydown.offcanvas.amui', function(e) {
    (e.keyCode === 27) && _this.close();
  });
};

OffCanvas.prototype.close = function(relatedElement) {
  var _this = this;
  var $html = $('html');
  var $body = $('body');
  var $element = this.$element;
  var $bar = $element.find('.am-offcanvas-bar').first();

  if (!$element.length || !this.active || !$element.hasClass('am-active')) {
    return;
  }

  $element.trigger('close.offcanvas.amui');

  function complete() {
    $body
      .removeClass('am-offcanvas-page')
      .css({
        width: '',
        height: '',
        'margin-left': '',
        'margin-right': ''
      });
    $element.removeClass('am-active');
    $bar.removeClass('am-offcanvas-bar-active');
    $html.css('margin-top', '');
    window.scrollTo(scrollPos.x, scrollPos.y);
    $element.trigger('closed.offcanvas.amui');
    _this.active = 0;
  }

  if (UI.support.transition) {
    setTimeout(function() {
      $bar.removeClass('am-offcanvas-bar-active');
    }, 0);

    $body.css('margin-left', '').one(UI.support.transition.end, function() {
      complete();
    }).emulateTransitionEnd(this.options.duration);
  } else {
    complete();
  }

  $element.off('click.offcanvas.amui');
  $html.off('.offcanvas.amui');
};

OffCanvas.prototype.bindEvents = function() {
  var _this = this;
  $doc.on('click.offcanvas.amui', '[data-am-dismiss="offcanvas"]', function(e) {
      e.preventDefault();
      _this.close();
    });

  $win.on('resize.offcanvas.amui orientationchange.offcanvas.amui',
    function() {
      _this.active && _this.close();
    });

  this.$element.hammer().on('swipeleft swipeleft', function(e) {
    e.preventDefault();
    _this.close();
  });

  return this;
};

function Plugin(option, relatedElement) {
  var args = Array.prototype.slice.call(arguments, 1);

  return this.each(function() {
    var $this = $(this);
    var data = $this.data('amui.offcanvas');
    var options = $.extend({}, typeof option == 'object' && option);

    if (!data) {
      $this.data('amui.offcanvas', (data = new OffCanvas(this, options)));
      (!option || typeof option == 'object') && data.open(relatedElement);
    }

    if (typeof option == 'string') {
      data[option] && data[option].apply(data, args);
    }
  });
}

$.fn.offCanvas = Plugin;

// Init code
$doc.on('click.offcanvas.amui', '[data-am-offcanvas]', function(e) {
  e.preventDefault();
  var $this = $(this);
  var options = UI.utils.parseOptions($this.data('amOffcanvas'));
  var $target = $(options.target ||
  (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
  var option = $target.data('amui.offcanvas') ? 'open' : options;

  Plugin.call($target, option, this);
});

module.exports = UI.offcanvas = OffCanvas;

// TODO: 优化动画效果
// http://dbushell.github.io/Responsive-Off-Canvas-Menu/step4.html
