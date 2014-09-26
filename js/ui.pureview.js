define(function(require, exports, module) {

    require('core');

    var PinchZoom = require('zepto.pinchzoom'),
        Hammer = require('util.hammer'),
        $ = window.Zepto,
        UI = $.AMUI,
        animation = UI.support.animation,
        transition = UI.support.transition,
        $html = $('html');

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
        this.$pureview = $(this.options.tpl, {
            id: UI.utils.generateGUID('am-pureview')
        });

        this.$slides = null;
        this.transitioning = null;
        this.scrollbarWidth = 0;

        this.init();
    };

    PureView.DEFAULTS = {
        tpl: '<div class="am-pureview am-pureview-bar-active">' +
        '<ul class="am-pureview-slider"></ul>' +
        '<ul class="am-pureview-direction"><li class="am-pureview-prev"><a href=""></a></li><li class="am-pureview-next"><a href=""></a></li></ul>' +
        '<ol class="am-pureview-nav"></ol>' +
        '<div class="am-pureview-bar am-active"><span class="am-pureview-title"></span><span class="am-pureview-current"></span> / <span class="am-pureview-total"></span></div>' +
        '<div class="am-pureview-actions am-active"><a href="javascript: void(0)" class="am-icon-chevron-left" data-am-close="pureview"></a></div>' +
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

        // 从何处获取图片，img 可以使用 data-rel 指定大图
        target: 'img',

        // 微信 Webview 中调用微信的图片浏览器
        // 实现图片保存、分享好友、收藏图片等功能
        weChatImagePreview: true
    };

    PureView.prototype.init = function() {
        var me = this,
            options = this.options,
            $element = this.$element,
            $pureview = this.$pureview,
            $slider = $pureview.find(options.selector.slider),
            $nav = $pureview.find(options.selector.nav),
            $slides = $([]),
            $navItems = $([]),
            $images = $element.find(options.target),
            total = $images.length,
            imgUrls = [];

        if (!total) return;

        if (total === 1) {
            $pureview.addClass(options.className.onlyOne);
        }

        $images.each(function(i, item) {
            var src, title;

            if (options.target == 'a') {
                src = item.href; // to absolute path
                title = item.title || ''
            } else {
                src = $(item).data('rel') || item.src; // <img src='' data-rel='' />
                title = $(item).attr('alt') || '';
            }

            // hide bar: wechat_webview_type=1
            // http://tmt.io/wechat/  not working?
            imgUrls.push(src);

            $slides = $slides.add($('<li><div class="am-pinch-zoom"><img src="' + src + '" alt="' + title + '"/></div></li>'));
            $navItems = $navItems.add($('<li>' + (i + 1) + '</li>'));
        });

        $slider.append($slides);
        $nav.append($navItems);

        $('body').append($pureview);

        $pureview.find(options.selector.total).text(total);

        this.$title = $pureview.find(options.selector.title);
        this.$current = $pureview.find(options.selector.current);
        this.$bar = $pureview.find(options.selector.bar);
        this.$actions = $pureview.find(options.selector.actions);
        this.$navItems = $nav.find('li');
        this.$slides = $slider.find('li');

        if (options.shareBtn) {
            this.$actions.append('<a href="javascript: void(0)" class="am-icon-share-square-o" data-am-toggle="share"></a>');
        }

        $slider.find(options.selector.pinchZoom).each(function() {
            $(this).data('amui.pinchzoom', new PinchZoom($(this), {}));
            $(this).on('pz_doubletap', function(e) {
                //
            });
        });

        $images.on('click.pureview.amui', function(e) {
            e.preventDefault();
            var clicked = $images.index(this);

            // Invoke WeChat ImagePreview in WeChat
            // TODO: detect WeChat before init
            if (options.weChatImagePreview && window.WeixinJSBridge) {
                WeixinJSBridge.invoke('imagePreview', {
                    'current' : imgUrls[clicked],
                    'urls' : imgUrls
                });
            } else {
                me.open(clicked);
            }
        });

        $pureview.find('.am-pureview-direction a').on('click.direction.pureview.amui', function(e) {
            e.preventDefault();
            var $clicked = $(e.target).parent('li');

            if ($clicked.is('.am-pureview-prev')) {
                me.prevSlide();
            } else {
                me.nextSlide();
            }
        });

        // Nav Contorl
        this.$navItems.on('click.nav.pureview.amui', function() {
            var index = me.$navItems.index($(this));
            me.activate(me.$slides.eq(index));
        });

        // Close Icon
        $pureview.find(options.selector.close).on('click.close.pureview.amui', function(e) {
            e.preventDefault();
            me.close();
        });

        $slider.hammer().on('press.pureview.amui', function(e) {
            e.preventDefault();
            me.toggleToolBar();
        }).on('swipeleft.pureview.amui', function(e) {
            e.preventDefault();
            me.nextSlide();
        }).on('swiperight.pureview.amui', function(e) {
            e.preventDefault();
            me.prevSlide();
        });

        $slider.data('hammer').get('swipe').set({
            direction: Hammer.DIRECTION_HORIZONTAL,
            velocity: 0.35
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

    PureView.prototype.activate = function($slide) {
        var options = this.options,
            $slides = this.$slides,
            activeIndex = $slides.index($slide),
            alt = $slide.find('img').attr('alt') || '',
            active = options.className.active;
        
        UI.utils.imageLoader($slide.find('img'), function(image) {
            $(image).addClass('am-img-loaded');
        });

        if ($slides.find('.' + active).is($slide)) return;

        if (this.transitioning) return;

        this.transitioning = 1;

        this.$title.text(alt);
        this.$current.text(activeIndex + 1);
        $slides.removeClass();
        $slide.addClass(active);
        $slides.eq(activeIndex - 1).addClass(options.className.prevSlide);
        $slides.eq(activeIndex + 1).addClass(options.className.nextSlide);

        this.$navItems.removeClass().eq(activeIndex).addClass(options.className.active);

        if (transition) {
            $slide.one(transition.end, $.proxy(function() {
                this.transitioning = 0
            }, this));
        } else {
            this.transitioning = 0
        }
    };

    PureView.prototype.nextSlide = function() {
        if (this.$slides.length === 1) return;

        var $slides = this.$slides,
            $active = $slides.filter('.am-active'),
            activeIndex = $slides.index($active),
            rightSpring = 'am-animation-right-spring';

        if (activeIndex + 1 >= $slides.length) { // last one
            animation && $active.addClass(rightSpring).on(animation.end, function() {
                $active.removeClass(rightSpring);
            });
        } else {
            this.activate($slides.eq(activeIndex + 1));
        }
    };

    PureView.prototype.prevSlide = function() {
        if (this.$slides.length === 1) return;

        var $slides = this.$slides,
            $active = $slides.filter('.am-active'),
            activeIndex = this.$slides.index(($active)),
            leftSpring = 'am-animation-left-spring';

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
        this.$pureview.addClass(this.options.className.active);
        this.$body.addClass(this.options.className.activeBody)
    };

    PureView.prototype.close = function() {
        var options = this.options;

        this.$pureview.removeClass(options.className.active);
        this.$slides.removeClass();

        function resetBody() {
            this.$body.removeClass(options.className.activeBody);
            this.resetScrollbar();
        }

        if (transition) {
            this.$pureview.one(transition.end, $.proxy(resetBody, this));
        } else {
            resetBody.call(this);
        } 
    };

    PureView.prototype.checkScrollbar = function () {
        this.scrollbarWidth = UI.utils.measureScrollbar();
    };

    PureView.prototype.setScrollbar = function () {
        var bodyPaddingRight = parseInt((this.$body.css('padding-right') || 0), 10);
        if (this.scrollbarWidth) this.$body.css('padding-right', bodyPaddingRight + this.scrollbarWidth);
    };

    PureView.prototype.resetScrollbar = function () {
        this.$body.css('padding-right', '');
    };

    UI.pureview = PureView;

    function Plugin(option) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.pureview'),
                options = $.extend({}, UI.utils.parseOptions($this.data('amPureview')), typeof option == 'object' && option);

            if (!data) {
                $this.data('am.pureview', (data = new PureView(this, options)));
            }

            if (typeof option == 'string') {
                data[option]();
            }
        });
    }

    $.fn.pureview = Plugin;

    // Init code
    $(function() {
        $('[data-am-pureview]').pureview();
    });

    module.exports = PureView;
});

// TODO: 1. 动画改进
//       2. 改变图片的时候恢复 Zoom
//       3. 选项
//       4. 图片高度问题：由于 PinchZoom 的原因，过高的图片如果设置看了滚动，则放大以后显示不全