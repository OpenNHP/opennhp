'use strict';

var $ = require('jquery');
var UI = require('./core');
var Hammer = require('./util.hammer');
var supportTransition = UI.support.transition;
var animation = UI.support.animation;

/**
 * @via https://github.com/twbs/bootstrap/blob/master/js/tab.js
 * @copyright 2011-2014 Twitter, Inc.
 * @license MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE)
 */

/**
 * Tabs
 * @param {HTMLElement} element
 * @param {Object} options
 * @constructor
 */
var Tabs = function(element, options) {
  this.$element = $(element);
  this.options = $.extend({}, Tabs.DEFAULTS, options || {});
  this.transitioning = this.activeIndex = null;

  this.refresh();
  this.init();
};

Tabs.DEFAULTS = {
  selector: {
    nav: '> .am-tabs-nav',
    content: '> .am-tabs-bd',
    panel: '> .am-tab-panel'
  },
  activeClass: 'am-active'
};

Tabs.prototype.refresh = function() {
  var selector = this.options.selector;

  this.$tabNav = this.$element.find(selector.nav);
  this.$navs = this.$tabNav.find('a');

  this.$content = this.$element.find(selector.content);
  this.$tabPanels = this.$content.find(selector.panel);

  var $active = this.$tabNav.find('> .' + this.options.activeClass);

  // Activate the first Tab when no active Tab or multiple active Tabs
  if ($active.length !== 1) {
    this.open(0);
  } else {
    this.activeIndex = this.$navs.index($active.children('a'));
  }
};

Tabs.prototype.init = function() {
  var _this = this;
  var options = this.options;

  this.$element.on('click.tabs.amui', options.selector.nav + ' a', function(e) {
    e.preventDefault();
    _this.open($(this));
  });

  // TODO: nested Tabs touch events
  if (!options.noSwipe) {
    if (!this.$content.length) {
      return this;
    }

    var hammer = new Hammer.Manager(this.$content[0]);
    var swipe = new Hammer.Swipe({
      direction: Hammer.DIRECTION_HORIZONTAL
      // threshold: 40
    });

    hammer.add(swipe);

    hammer.on('swipeleft', UI.utils.debounce(function(e) {
      e.preventDefault();
      _this.goTo('next');
    }, 100));

    hammer.on('swiperight', UI.utils.debounce(function(e) {
      e.preventDefault();
      _this.goTo('prev');
    }, 100));

    this._hammer = hammer;
  }
};

/**
 * Open $nav tab
 * @param {jQuery|HTMLElement|Number} $nav
 * @returns {Tabs}
 */
Tabs.prototype.open = function($nav) {
  var activeClass = this.options.activeClass;
  var activeIndex = typeof $nav === 'number' ? $nav : this.$navs.index($($nav));

  $nav = typeof $nav === 'number' ? this.$navs.eq(activeIndex) : $($nav);

  if (!$nav ||
    !$nav.length ||
    this.transitioning ||
    $nav.parent('li').hasClass(activeClass)) {
    return;
  }

  var $tabNav = this.$tabNav;
  var href = $nav.attr('href');
  var regexHash = /^#.+$/;
  var $target = regexHash.test(href) && this.$content.find(href) ||
    this.$tabPanels.eq(activeIndex);
  var previous = $tabNav.find('.' + activeClass + ' a')[0];
  var e = $.Event('open.tabs.amui', {
    relatedTarget: previous
  });

  $nav.trigger(e);

  if (e.isDefaultPrevented()) {
    return;
  }

  // activate Tab nav
  this.activate($nav.closest('li'), $tabNav);

  // activate Tab content
  this.activate($target, this.$content, function() {
    $nav.trigger({
      type: 'opened.tabs.amui',
      relatedTarget: previous
    });
  });

  this.activeIndex = activeIndex;
};

Tabs.prototype.activate = function($element, $container, callback) {
  this.transitioning = true;

  var activeClass = this.options.activeClass;
  var $active = $container.find('> .' + activeClass);
  var transition = callback && supportTransition && !!$active.length;

  $active.removeClass(activeClass + ' am-in');

  $element.addClass(activeClass);

  if (transition) {
    $element.redraw(); // reflow for transition
    $element.addClass('am-in');
  } else {
    $element.removeClass('am-fade');
  }

  var complete = $.proxy(function complete() {
    callback && callback();
    this.transitioning = false;
  }, this);



  transition && !this.$content.is('.am-tabs-bd-ofv') ?
    $active.one(supportTransition.end, complete) : complete();
};

/**
 * Go to `next` or `prev` tab
 * @param {String} direction - `next` or `prev`
 */
Tabs.prototype.goTo = function(direction) {
  var navIndex = this.activeIndex;
  var isNext = direction === 'next';
  var spring = isNext ? 'am-animation-right-spring' :
    'am-animation-left-spring';

  if ((isNext && navIndex + 1 >= this.$navs.length) || // last one
    (!isNext && navIndex === 0)) { // first one
    var $panel = this.$tabPanels.eq(navIndex);

    animation && $panel.addClass(spring).on(animation.end, function() {
      $panel.removeClass(spring);
    });
  } else {
    this.open(isNext ? navIndex + 1 : navIndex - 1);
  }
};

Tabs.prototype.destroy = function() {
  this.$element.off('.tabs.amui');
  Hammer.off(this.$content[0], 'swipeleft swiperight');
  this._hammer && this._hammer.destroy();
  $.removeData(this.$element, 'amui.tabs');
};

// Plugin
function Plugin(option) {
  var args = Array.prototype.slice.call(arguments, 1);
  var methodReturn;

  this.each(function() {
    var $this = $(this);
    var $tabs = $this.is('.am-tabs') && $this || $this.closest('.am-tabs');
    var data = $tabs.data('amui.tabs');
    var options = $.extend({}, UI.utils.parseOptions($this.data('amTabs')),
      $.isPlainObject(option) && option);

    if (!data) {
      $tabs.data('amui.tabs', (data = new Tabs($tabs[0], options)));
    }

    if (typeof option === 'string') {
      if (option === 'open' && $this.is('.am-tabs-nav a')) {
        data.open($this);
      } else {
        methodReturn = typeof data[option] === 'function' ?
          data[option].apply(data, args) : data[option];
      }
    }
  });

  return methodReturn === undefined ? this : methodReturn;
}

$.fn.tabs = Plugin;

// Init code
UI.ready(function(context) {
  $('[data-am-tabs]', context).tabs();
});

$(document).on('click.tabs.amui.data-api', '[data-am-tabs] .am-tabs-nav a',
  function(e) {
  e.preventDefault();
  Plugin.call($(this), 'open');
});

module.exports = UI.tabs = Tabs;

// TODO: 1. Ajax 支持
//       2. touch 事件处理逻辑优化
