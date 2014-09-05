define(function(require, exports, module) {

    require('core');

    var PinchZoom = require('zepto.pinchzoom'),
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
        tpl: '<div class="am-pureview">' +
        '<ul class="am-pureview-slider"></ul>' +
        '<ul class="am-pureview-direction"><li class="am-pureview-prev"><a href=""></a></li><li class="am-pureview-next"><a href=""></a></li></ul>' +
        '<div class="am-pureview-bar am-active"><span class="am-pureview-current"></span> / <span class="am-pureview-total"></span><span class="am-pureview-title"></span></div>' +
        '<div class="am-pureview-actions am-active"><a href="javascript: void(0)" class="am-icon-chevron-left" data-am-close="pureview"></a><a href="javascript: void(0)" class="am-icon-share-square-o" data-am-toggle="share"></a></div>' +
        '</div>',

        className: {
            prevSlide: 'am-pureview-slide-prev',
            nextSlide: 'am-pureview-slide-next',
            active: 'am-active'
        },

        selector: {
            slider: '.am-pureview-slider',
            close: '[data-am-close="pureview"]',
            total: '.am-pureview-total',
            current: '.am-pureview-current',
            title: '.am-pureview-title',
            actions: '.am-pureview-actions',
            bar: '.am-pureview-bar',
            pinchZoom: '.am-pinch-zoom'
        }
    };

    PureView.prototype.init = function() {
        var me = this,
            options = me.options,
            $element = me.$element,
            $images = $element.find('img'),
            $pureview = me.$pureview,
            $slider = $pureview.find(options.selector.slider),
            slides = [],
            total = $images.length;

        if (!total) return;

        $images.each(function(i, img) {
            var alt = $(img).attr('alt') || '',
                slide = '<li><div class="am-pinch-zoom"><img src="' + img.src +'" alt="' + alt + '"/></div></li>';
            slides.push(slide);
        });

        $slider.html(slides.join('\n'));

        $('body').append($pureview);

        $pureview.find(options.selector.total).text(total);

        this.$title = $pureview.find(options.selector.title);
        this.$current = $pureview.find(options.selector.current);
        this.$bar = $pureview.find(options.selector.bar);
        this.$actions = $pureview.find(options.selector.actions);

        this.$slides = $slider.find('li');

        $slider.find(options.selector.pinchZoom).each(function() {
            $(this).data('amui.pinchzoom', new PinchZoom($(this), {}));
        });

        $images.on('click', function(e) {
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

        // Close Icon
        $pureview.find(options.selector.close).on('click.pureview.amui', function(e) {
            e.preventDefault();
            me.close();
        });

        $slider.on('singleTap', function(e) {
            me.toggleToolBar();
        }).on('swipeLeft', function(e) {
            me.nextSlide()
        }).on('swipeRight', function(e) {
            me.prevSlide();
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
        $slides.removeAttr('class');
        $slide.addClass(active);
        $slides.eq(activeIndex - 1).addClass(options.className.prevSlide);
        $slides.eq(activeIndex + 1).addClass(options.className.nextSlide);

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
        var active = this.options.className.active;
        this.$bar.toggleClass(active);
        this.$actions.toggleClass(active);
    };

    PureView.prototype.open = function(index) {
        var active = index || 0;
        this.activate(this.$slides.eq(active));
        this.$pureview.addClass('am-active');
        $html.addClass('am-dimmer-active')
    };

    PureView.prototype.close = function() {
        this.$pureview.removeClass('am-active');
        $html.removeClass('am-dimmer-active')
    };

    UI.pureview = PureView;

    function Plugin(option) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.pureview'),
                options = $.extend({}, UI.utils.parseOptions($this.attr('data-am-pureview')), typeof option == 'object' && option);
            
            console.log(data);

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

// TODO: 1. 动画优化
//       2. 替换触控动画库
