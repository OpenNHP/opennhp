define(function(require, exports, module) {
  'use strict';

  require('zepto.outerdemension');
  require('zepto.extend.data');
  require('core');

  var $ = window.Zepto;
  var UI = $.AMUI;
  var $win = $(window);
  var $doc = $(document);
  var scrollPos;

  /**
   * @via https://github.com/uikit/uikit/blob/master/src/js/offcanvas.js
   * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
   */

  var OffCanvas = function(element, options) {
    this.$element = $(element);
    this.options = options;
    this.active = null;
    this.events();
  };

  OffCanvas.DEFAULTS = {
    duration: 300,
    effect: 'overlay' // {push|overlay}, push is too expensive
  };

  OffCanvas.prototype.open = function(relatedElement) {
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

    $body.
        css({width: window.innerWidth, height: $win.height()}).
        addClass('am-offcanvas-page');

    if (effect !== 'overlay') {
      $body.css({
        'margin-left': $bar.outerWidth() * dir
      }).width(); // force redraw
    }

    $html.css('margin-top', scrollPos.y * -1);

    setTimeout(function() {
      $bar.addClass('am-offcanvas-bar-active').width();
    }, 0);

    $doc.trigger('open:offcanvas:amui');

    this.active = 1;

    $element.off('.offcanvas.amui').
        on('click.offcanvas.amui swipe.offcanvas.amui', $.proxy(function(e) {
          var $target = $(e.target);

          if (!e.type.match(/swipe/)) {
            if ($target.hasClass('am-offcanvas-bar')) {
              return;
            }

            if ($target.parents('.am-offcanvas-bar').first().length) {
              return;
            }
          }

          // https://developer.mozilla.org/zh-CN/docs/DOM/event.stopImmediatePropagation
          e.stopImmediatePropagation();

          this.close();
        }, this));

    $html.on('keydown.offcanvas.amui', $.proxy(function(e) {
      if (e.keyCode === 27) { // ESC
        this.close();
      }
    }, this));
  };

  OffCanvas.prototype.close = function(relatedElement) {
    var me = this;
    var $html = $('html');
    var $body = $('body');
    var $element = this.$element;
    var $bar = $element.find('.am-offcanvas-bar').first();

    if (!$element.length || !$element.hasClass('am-active')) {
      return;
    }

    $doc.trigger('close:offcanvas:amui');

    function complete() {
      $body.removeClass('am-offcanvas-page').
          css({width: '', height: '', 'margin-left': '', 'margin-right': ''});
      $element.removeClass('am-active');
      $bar.removeClass('am-offcanvas-bar-active');
      $html.css('margin-top', '');
      window.scrollTo(scrollPos.x, scrollPos.y);
      $doc.trigger('closed:offcanvas:amui');
      me.active = 0;
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

    $element.off('.offcanvas.amui');
    $html.off('.offcanvas.amui');
  };

  OffCanvas.prototype.events = function() {
    $doc.on('click.offcanvas.amui', '[data-am-dismiss="offcanvas"]',
        $.proxy(function(e) {
          e.preventDefault();
          this.close();
        }, this));

    $win.on('resize.offcanvas.amui orientationchange.offcanvas.amui',
        $.proxy(function(e) {
          this.active && this.close();
        }, this));

    return this;
  };

  function Plugin(option, relatedElement) {
    return this.each(function() {
      var $this = $(this);
      var data = $this.data('am.offcanvas');
      var options = $.extend({}, OffCanvas.DEFAULTS,
                  typeof option == 'object' && option);

      if (!data) {
        $this.data('am.offcanvas', (data = new OffCanvas(this, options)));
        data.open(relatedElement);
      }

      if (typeof option == 'string') {
        data[option] && data[option](relatedElement);
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
    var option = $target.data('am.offcanvas') ? 'open' : options;

    Plugin.call($target, option, this);
  });

  UI.offcanvas = OffCanvas;

  module.exports = OffCanvas;
});

// TODO: 优化动画效果
// http://dbushell.github.io/Responsive-Off-Canvas-Menu/step4.html
