define(function(require, exports, module) {

    require('core');

    var PinchZoom = require('zepto.pinchzoom'),
        Hammer = require('util.hammer'),
        $ = window.Zepto,
        UI = $.AMUI,
        animation = UI.support.animation,
        transition = UI.support.transition,
        $html = $('html');

    var PureView = function(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, PureView.DEFAULTS, options);
        this.$pureview = $(this.options.tpl, {
            id: UI.utils.generateGUID('am-pureview')
        });

        this.$slides = null;
        this.transitioning = null;

        this.init();
    };

    PureView.DEFAULTS = {
        tpl: '<div class="am-pureview am-pureview-bar-active">' +
        '<ul class="am-pureview-slider"></ul>' +
        '<ul class="am-pureview-direction"><li class="am-pureview-prev"><a href=""></a></li><li class="am-pureview-next"><a href=""></a></li></ul>' +
        '<ol class="am-pureview-nav"></ol>' +
        '<div class="am-pureview-bar am-active"><span class="am-pureview-current"></span> / <span class="am-pureview-total"></span><span class="am-pureview-title"></span></div>' +
        '<div class="am-pureview-actions am-active"><a href="javascript: void(0)" class="am-icon-chevron-left" data-am-close="pureview"></a><a href="javascript: void(0)" class="am-icon-share-square-o" data-am-toggle="share"></a></div>' +
        '</div>',

        className: {
            prevSlide: 'am-pureview-slide-prev',
            nextSlide: 'am-pureview-slide-next',
            active: 'am-active',
            barActive: 'am-pureview-bar-active',
            onlyOne: 'am-pureview-only'
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
        }
    };

    PureView.prototype.init = function() {
        var me = this,
            options = me.options,
            $element = me.$element,
            $images = $element.find('img'),
            $pureview = me.$pureview,
            $slider = $pureview.find(options.selector.slider),
            $nav = $pureview.find(options.selector.nav),
            $slides = $([]),
            total = $images.length,
            $navItems = $([]);

        if (!total) return;

        if (total === 1) {
            $pureview.addClass(options.className.onlyOne);
        }

        $images.each(function(i, img) {
            var alt = $(img).attr('alt') || '';

            $slides = $slides.add($('<li><div class="am-pinch-zoom"><img src="' + img.src + '" alt="' + alt + '"/></div></li>'));
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

        $slider.find(options.selector.pinchZoom).each(function() {
            $(this).data('amui.pinchzoom', new PinchZoom($(this), {}));
            $(this).on('pz_doubletap', function(e) {
                //
            });
        });

        $images.on('click.pureview.amui', function(e) {
            e.preventDefault();
            me.open($images.index(this));
        });

        $pureview.find('.am-pureview-direction a').on('click', function(e) {
            e.preventDefault();
            var $clicked = $(e.target).parent('li');

            if ($clicked.is('.am-pureview-prev')) {
                me.prevSlide();
            } else {
                me.nextSlide();
            }
        });

        // Nav Contorl
        this.$navItems.on('click.pureview.amui', function() {
            var index = me.$navItems.index($(this));
            me.activate(me.$slides.eq(index));
        });

        // Close Icon
        $pureview.find(options.selector.close).on('click.pureview.amui', function(e) {
            e.preventDefault();
            me.close();
        });

        $slider.hammer().on('press.pureview.amui', function(e) {
            me.toggleToolBar();
        }).on('swipeleft.pureview.amui', function(e) {
            me.nextSlide();
        }).on('swiperight.pureview.amui', function(e) {
            me.prevSlide();
        });

        $slider.data('hammer').get('swipe').set({
            direction: Hammer.DIRECTION_HORIZONTAL,
            threshold: 50,
            velocity: 0.45
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
            alt = $slide.find('img').attr('alt'),
            active = options.className.active;

        if ($slides.find('.' + active).is($slide)) return;

        if (this.transitioning) return;

        this.transitioning = 1;

        alt && this.$title.text(alt);
        this.$current.text(activeIndex + 1);
        $slides.removeClass();
        $slide.addClass(active);
        $slides.eq(activeIndex - 1).addClass(options.className.prevSlide);
        $slides.eq(activeIndex + 1).addClass(options.className.nextSlide);

        this.$navItems.removeClass().eq(activeIndex).addClass('am-active');

        if (transition) {
            $slide.one(transition.end, $.proxy(function() {
                this.transitioning = 0
            }, this));
        } else {
            this.transitioning = 0
        }
    };

    PureView.prototype.nextSlide = function() {
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
        this.activate(this.$slides.eq(active));
        this.$pureview.addClass('am-active');
        $html.addClass('am-dimmer-active')
    };

    PureView.prototype.close = function() {
        this.$pureview.removeClass('am-active');
        this.$slides.removeClass();
        
        if (transition) {
            this.$pureview.one(transition.end, function() {
                $html.removeClass('am-dimmer-active');
            });
        } else {
            $html.removeClass('am-dimmer-active');
        } 
    };

    UI.pureview = PureView;

    function Plugin(option) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.pureview'),
                options = $.extend({}, UI.utils.parseOptions($this.attr('data-am-pureview')), typeof option == 'object' && option);

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
//       4. 关闭以后滚动条位置处理