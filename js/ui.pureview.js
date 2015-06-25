'use strict';

var $ = require('jquery');
var UI = require('./core');
var PinchZoom = require('./ui.pinchzoom');
var Hammer = require('./util.hammer');
var animation = UI.support.animation;
var transition = UI.support.transition;

/**
 * PureView
 * @desc Image browser for Mobile
 * @param element
 * @param options
 * @constructor
 */

var PureView = function(element, options) {
  this.$element = $(element);
  this.$body = $(document.body);
  this.options = $.extend({}, PureView.DEFAULTS, options);
  this.$pureview = $(this.options.tpl).attr('id',
    UI.utils.generateGUID('am-pureview'));

  this.$slides = null;
  this.transitioning = null;
  this.scrollbarWidth = 0;

  this.init();
};

PureView.DEFAULTS = {
  tpl: '<div class="am-pureview am-pureview-bar-active">' +
  '<ul class="am-pureview-slider"></ul>' +
  '<ul class="am-pureview-direction">' +
  '<li class="am-pureview-prev"><a href=""></a></li>' +
  '<li class="am-pureview-next"><a href=""></a></li></ul>' +
  '<ol class="am-pureview-nav"></ol>' +
  '<div class="am-pureview-bar am-active">' +
  '<span class="am-pureview-title"></span>' +
  '<div class="am-pureview-counter"><span class="am-pureview-current"></span> / ' +
  '<span class="am-pureview-total"></span></div></div>' +
  '<div class="am-pureview-actions am-active">' +
  '<a href="javascript: void(0)" class="am-icon-chevron-left" ' +
  'data-am-close="pureview"></a></div>' +
  '</div>',

  className: {
    prevSlide: 'am-pureview-slide-prev',
    nextSlide: 'am-pureview-slide-next',
    onlyOne: 'am-pureview-only',
    active: 'am-active',
    barActive: 'am-pureview-bar-active',
    activeBody: 'am-pureview-active'
  },

  selector: {
    slider: '.am-pureview-slider',
    close: '[data-am-close="pureview"]',
    total: '.am-pureview-total',
    current: '.am-pureview-current',
    title: '.am-pureview-title',
    actions: '.am-pureview-actions',
    bar: '.am-pureview-bar',
    pinchZoom: '.am-pinch-zoom',
    nav: '.am-pureview-nav'
  },

  shareBtn: false,

  // press to toggle Toolbar
  toggleToolbar: true,

  // 从何处获取图片，img 可以使用 data-rel 指定大图
  target: 'img',

  // 微信 Webview 中调用微信的图片浏览器
  // 实现图片保存、分享好友、收藏图片等功能
  weChatImagePreview: true
};

PureView.prototype.init = function() {
  var _this = this;
  var options = this.options;
  var $element = this.$element;
  var $pureview = this.$pureview;

  this.refreshSlides();

  $('body').append($pureview);

  this.$title = $pureview.find(options.selector.title);
  this.$current = $pureview.find(options.selector.current);
  this.$bar = $pureview.find(options.selector.bar);
  this.$actions = $pureview.find(options.selector.actions);

  if (options.shareBtn) {
    this.$actions.append('<a href="javascript: void(0)" ' +
    'class="am-icon-share-square-o" data-am-toggle="share"></a>');
  }

  this.$element.on('click.pureview.amui', options.target, function(e) {
    e.preventDefault();
    var clicked = _this.$images.index(this);

    // Invoke WeChat ImagePreview in WeChat
    // TODO: detect WeChat before init
    if (options.weChatImagePreview && window.WeixinJSBridge) {
      window.WeixinJSBridge.invoke('imagePreview', {
        current: _this.imgUrls[clicked],
        urls: _this.imgUrls
      });
    } else {
      _this.open(clicked);
    }
  });

  $pureview.find('.am-pureview-direction').
    on('click.direction.pureview.amui', 'li', function(e) {
      e.preventDefault();

      if ($(this).is('.am-pureview-prev')) {
        _this.prevSlide();
      } else {
        _this.nextSlide();
      }
    });

  // Nav Contorl
  $pureview.find(options.selector.nav).on('click.nav.pureview.amui', 'li',
    function() {
      var index = _this.$navItems.index($(this));
      _this.activate(_this.$slides.eq(index));
    });

  // Close Icon
  $pureview.find(options.selector.close).
    on('click.close.pureview.amui', function(e) {
      e.preventDefault();
      _this.close();
    });

  this.$slider.hammer().on('swipeleft.pureview.amui', function(e) {
    e.preventDefault();
    _this.nextSlide();
  }).on('swiperight.pureview.amui', function(e) {
    e.preventDefault();
    _this.prevSlide();
  }).on('press.pureview.amui', function(e) {
    e.preventDefault();
    options.toggleToolbar && _this.toggleToolBar();
  });

  this.$slider.data('hammer').get('swipe').set({
    direction: Hammer.DIRECTION_HORIZONTAL,
    velocity: 0.35
  });

  // Observe DOM
  $element.DOMObserve({
    childList: true,
    subtree: true
  }, function(mutations, observer) {
    // _this.refreshSlides();
    // console.log('mutations[0].type);
  });

  // NOTE:
  // trigger this event manually if MutationObserver not supported
  //   when new images appended, or call refreshSlides()
  // if (!UI.support.mutationobserver) $element.trigger('changed.dom.amui')
  $element.on('changed.dom.amui', function(e) {
    e.stopPropagation();
    _this.refreshSlides();
  });

  $(document).on('keydown.pureview.amui', $.proxy(function(e) {
    var keyCode = e.keyCode;
    if (keyCode == 37) {
      this.prevSlide();
    } else if (keyCode == 39) {
      this.nextSlide();
    } else if (keyCode == 27) {
      this.close();
    }
  }, this));
};

PureView.prototype.refreshSlides = function() {
  // update images collections
  this.$images = this.$element.find(this.options.target);
  var _this = this;
  var options = this.options;
  var $pureview = this.$pureview;
  var $slides = $([]);
  var $navItems = $([]);
  var $images = this.$images;
  var total = $images.length;
  this.$slider = $pureview.find(options.selector.slider);
  this.$nav = $pureview.find(options.selector.nav);
  var viewedFlag = 'data-am-pureviewed';
  // for WeChat Image Preview
  this.imgUrls = this.imgUrls || [];

  if (!total) {
    return;
  }

  if (total === 1) {
    $pureview.addClass(options.className.onlyOne);
  }

  $images.not('[' + viewedFlag + ']').each(function(i, item) {
    var src;
    var title;

    // get image URI from link's href attribute
    if (item.nodeName === 'A') {
      src = item.href; // to absolute path
      title = item.title || '';
    } else {
      // NOTE: `data-rel` should be a full URL, otherwise,
      //        WeChat images preview will not work
      src = $(item).data('rel') || item.src; // <img src='' data-rel='' />
      title = $(item).attr('alt') || '';
    }

    // add pureviewed flag
    item.setAttribute(viewedFlag, '1');

    // hide bar: wechat_webview_type=1
    // http://tmt.io/wechat/  not working?
    _this.imgUrls.push(src);

    $slides = $slides.add($('<li data-src="' + src + '" data-title="' + title +
    '"></li>'));
    $navItems = $navItems.add($('<li>' + (i + 1) + '</li>'));
  });

  $pureview.find(options.selector.total).text(total);

  this.$slider.append($slides);
  this.$nav.append($navItems);
  this.$navItems = this.$nav.find('li');
  this.$slides = this.$slider.find('li');
};

PureView.prototype.loadImage = function($slide, callback) {
  var appendedFlag = 'image-appended';

  if (!$slide.data(appendedFlag)) {
    var $img = $('<img>', {
      src: $slide.data('src'),
      alt: $slide.data('title')
    });

    $slide.html($img).wrapInner('<div class="am-pinch-zoom"></div>').redraw();

    var $pinchWrapper = $slide.find(this.options.selector.pinchZoom);
    $pinchWrapper.data('amui.pinchzoom', new PinchZoom($pinchWrapper[0], {}));
    $slide.data('image-appended', true);
  }

  callback && callback.call(this);
};

PureView.prototype.activate = function($slide) {
  var options = this.options;
  var $slides = this.$slides;
  var activeIndex = $slides.index($slide);
  var title = $slide.data('title') || '';
  var active = options.className.active;

  if ($slides.find('.' + active).is($slide)) {
    return;
  }

  if (this.transitioning) {
    return;
  }

  this.loadImage($slide, function() {
    UI.utils.imageLoader($slide.find('img'), function(image) {
      $slide.find('.am-pinch-zoom').addClass('am-pureview-loaded');
      $(image).addClass('am-img-loaded');
    });
  });

  this.transitioning = 1;

  this.$title.text(title);
  this.$current.text(activeIndex + 1);
  $slides.removeClass();
  $slide.addClass(active);
  $slides.eq(activeIndex - 1).addClass(options.className.prevSlide);
  $slides.eq(activeIndex + 1).addClass(options.className.nextSlide);

  this.$navItems.removeClass().
    eq(activeIndex).addClass(options.className.active);

  if (transition) {
    $slide.one(transition.end, $.proxy(function() {
      this.transitioning = 0;
    }, this)).emulateTransitionEnd(300);
  } else {
    this.transitioning = 0;
  }

  // TODO: pre-load next image
};

PureView.prototype.nextSlide = function() {
  if (this.$slides.length === 1) {
    return;
  }

  var $slides = this.$slides;
  var $active = $slides.filter('.am-active');
  var activeIndex = $slides.index($active);
  var rightSpring = 'am-animation-right-spring';

  if (activeIndex + 1 >= $slides.length) { // last one
    animation && $active.addClass(rightSpring).on(animation.end, function() {
      $active.removeClass(rightSpring);
    });
  } else {
    this.activate($slides.eq(activeIndex + 1));
  }
};

PureView.prototype.prevSlide = function() {
  if (this.$slides.length === 1) {
    return;
  }

  var $slides = this.$slides;
  var $active = $slides.filter('.am-active');
  var activeIndex = this.$slides.index(($active));
  var leftSpring = 'am-animation-left-spring';

  if (activeIndex === 0) { // first one
    animation && $active.addClass(leftSpring).on(animation.end, function() {
      $active.removeClass(leftSpring);
    });
  } else {
    this.activate($slides.eq(activeIndex - 1));
  }
};

PureView.prototype.toggleToolBar = function() {
  this.$pureview.toggleClass(this.options.className.barActive);
};

PureView.prototype.open = function(index) {
  var active = index || 0;
  this.checkScrollbar();
  this.setScrollbar();
  this.activate(this.$slides.eq(active));
  this.$pureview.show().redraw().addClass(this.options.className.active);
  this.$body.addClass(this.options.className.activeBody);
};

PureView.prototype.close = function() {
  var options = this.options;

  this.$pureview.removeClass(options.className.active);
  this.$slides.removeClass();

  function resetBody() {
    this.$pureview.hide();
    this.$body.removeClass(options.className.activeBody);
    this.resetScrollbar();
  }

  if (transition) {
    this.$pureview.one(transition.end, $.proxy(resetBody, this)).
      emulateTransitionEnd(300);
  } else {
    resetBody.call(this);
  }
};

PureView.prototype.checkScrollbar = function() {
  this.scrollbarWidth = UI.utils.measureScrollbar();
};

PureView.prototype.setScrollbar = function() {
  var bodyPaddingRight = parseInt((this.$body.css('padding-right') || 0), 10);
  if (this.scrollbarWidth) {
    this.$body.css('padding-right', bodyPaddingRight + this.scrollbarWidth);
  }
};

PureView.prototype.resetScrollbar = function() {
  this.$body.css('padding-right', '');
};

UI.plugin('pureview', PureView);

// Init code
UI.ready(function(context) {
  $('[data-am-pureview]', context).pureview();
});

module.exports = PureView;

// TODO: 1. 动画改进
//       2. 改变图片的时候恢复 Zoom
//       3. 选项
//       4. 图片高度问题：由于 PinchZoom 的原因，过高的图片如果设置看了滚动，则放大以后显示不全
