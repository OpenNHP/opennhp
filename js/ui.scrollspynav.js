'use strict';

var $ = require('jquery');
var UI = require('./core');
require('./ui.smooth-scroll');

/**
 * @via https://github.com/uikit/uikit/
 * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
 */

// ScrollSpyNav Class
var ScrollSpyNav = function(element, options) {
  this.options = $.extend({}, ScrollSpyNav.DEFAULTS, options);
  this.$element = $(element);
  this.anchors = [];

  this.$links = this.$element.find('a[href^="#"]').each(function(i, link) {
    this.anchors.push($(link).attr('href'));
  }.bind(this));

  this.$targets = $(this.anchors.join(', '));

  var processRAF = function() {
    UI.utils.rAF.call(window, $.proxy(this.process, this));
  }.bind(this);

  this.$window = $(window).on('scroll.scrollspynav.amui', processRAF)
    .on('resize.scrollspynav.amui orientationchange.scrollspynav.amui',
    UI.utils.debounce(processRAF, 50));

  processRAF();
  this.scrollProcess();
};

ScrollSpyNav.DEFAULTS = {
  className: {
    active: 'am-active'
  },
  closest: false,
  smooth: true,
  offsetTop: 0
};

ScrollSpyNav.prototype.process = function() {
  var scrollTop = this.$window.scrollTop();
  var options = this.options;
  var inViews = [];
  var $links = this.$links;

  var $targets = this.$targets;

  $targets.each(function(i, target) {
    if (UI.utils.isInView(target, options)) {
      inViews.push(target);
    }
  });

  // console.log(inViews.length);

  if (inViews.length) {
    var $target;

    $.each(inViews, function(i, item) {
      if ($(item).offset().top >= scrollTop) {
        $target = $(item);
        return false; // break
      }
    });

    if (!$target) {
      return;
    }

    if (options.closest) {
      $links.closest(options.closest).removeClass(options.className.active);
      $links.filter('a[href="#' + $target.attr('id') + '"]').
        closest(options.closest).addClass(options.className.active);
    } else {
      $links.removeClass(options.className.active).
        filter('a[href="#' + $target.attr('id') + '"]').
        addClass(options.className.active);
    }
  }
};

ScrollSpyNav.prototype.scrollProcess = function() {
  var $links = this.$links;
  var options = this.options;

  // smoothScroll
  if (options.smooth && $.fn.smoothScroll) {
    $links.on('click', function(e) {
      e.preventDefault();

      var $this = $(this);
      var $target = $($this.attr('href'));

      if (!$target) {
        return;
      }

      var offsetTop = options.offsetTop &&
        !isNaN(parseInt(options.offsetTop)) && parseInt(options.offsetTop) || 0;

      $(window).smoothScroll({position: $target.offset().top - offsetTop});
    });
  }
};

// ScrollSpyNav Plugin
UI.plugin('scrollspynav', ScrollSpyNav);

// Init code
UI.ready(function(context) {
  $('[data-am-scrollspy-nav]', context).scrollspynav();
});

module.exports = ScrollSpyNav;

// TODO: 1. 算法改进
//       2. 多级菜单支持
//       3. smooth scroll pushState
