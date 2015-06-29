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
  this.transitioning = null;

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

  // Activate the first Tab when no active Tab or multiple active Tabs
  if (this.$tabNav.find('> .' + this.options.activeClass).length !== 1) {
    var $tabNav = this.$tabNav;
    this.activate($tabNav.children('li').first(), $tabNav);
    this.activate(this.$tabPanels.first(), this.$content);
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

    var hammer = new Hammer(this.$content[0]);

    hammer.get('pan').set({
      direction: Hammer.DIRECTION_HORIZONTAL,
      threshold: 120
    });

    hammer.on('panleft', UI.utils.debounce(function(e) {
      e.preventDefault();
      var $target = $(e.target);

      if (!$target.is(options.selector.panel)) {
        $target = $target.closest(options.selector.panel);
      }

      $target.focus();

      var $nav = _this.getNextNav($target);
      $nav && _this.open($nav);
    }, 100));

    hammer.on('panright', UI.utils.debounce(function(e) {
      e.preventDefault();

      var $target = $(e.target);

      if (!$target.is(options.selector.panel)) {
        $target = $target.closest(options.selector.panel);
      }

      var $nav = _this.getPrevNav($target);

      $nav && _this.open($nav);
    }, 100));
  }
};

/**
 * Open $nav tab
 * @param {jQuery|HTMLElement|Number} $nav
 * @returns {Tabs}
 */
Tabs.prototype.open = function($nav) {
  var activeClass = this.options.activeClass;
  $nav = typeof $nav === 'number' ? this.$navs.eq($nav) : $($nav);

  if (!$nav ||
    this.transitioning ||
    $nav.parent('li').hasClass(activeClass)) {
    return;
  }

  var $tabNav = this.$tabNav;
  var href = $nav.attr('href');
  var regexHash = /^#.+$/;
  var $target = regexHash.test(href) && this.$content.find(href) ||
    this.$tabPanels.eq(this.$navs.index($nav));
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

  transition ? $active.one(supportTransition.end, complete) : complete();
};

Tabs.prototype.getNextNav = function($panel) {
  var navIndex = this.$tabPanels.index($panel);
  var rightSpring = 'am-animation-right-spring';

  if (navIndex + 1 >= this.$navs.length) { // last one
    animation && $panel.addClass(rightSpring).on(animation.end, function() {
      $panel.removeClass(rightSpring);
    });
    return null;
  } else {
    return this.$navs.eq(navIndex + 1);
  }
};

Tabs.prototype.getPrevNav = function($panel) {
  var navIndex = this.$tabPanels.index($panel);
  var leftSpring = 'am-animation-left-spring';

  if (navIndex === 0) { // first one
    animation && $panel.addClass(leftSpring).on(animation.end, function() {
      $panel.removeClass(leftSpring);
    });
    return null;
  } else {
    return this.$navs.eq(navIndex - 1);
  }
};

Tabs.prototype.destroy = function() {
  this.$element.off('.tabs.amui');
  Hammer.off(this.$content[0], 'panleft panright');
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

/*
// Init code
UI.ready(function(context) {
  $('[data-am-tabs]', context).tabs();
});
*/

$(document).on('click.tabs.amui.data-api', '[data-am-tabs] .am-tabs-nav a',
  function(e) {
  e.preventDefault();
  Plugin.call($(this), 'open');
});

module.exports = UI.tabs = Tabs;

// TODO: 1. Ajax 支持
//       2. touch 事件处理逻辑优化
