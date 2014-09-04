define(function(require, exports, module) {

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI,
        animation = UI.support.animation;

    var PureView = function(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, PureView.DEFAULTS, options);
        this.$pureview = $(this.options.tpl, {
            id: UI.utils.generateGUID('am-pureview')
        });

        this.$slides = null;

        this.init();
    };

    PureView.DEFAULTS = {
        tpl: '<div class="am-pureview"><ul class="am-pureview-slides"></ul><ul class="am-pureview-direction"><li class="am-pureview-prev"><a href=""></a></li><li class="am-pureview-next"><a href=""></a></li></ul>' +
        '<div class="am-pureview-bar"><span class="am-pureview-current"></span> / <span class="am-pureview-total"></span><span class="am-pureview-title"></span></div>' +
        '<div class="am-pureview-actions"><a href="javascript: void(0)" class="am-icon-chevron-left" data-am-close="pureview"></a><a href="javascript: void(0)" class="am-icon-share-square-o" data-am-toggle="share"></a></div>' +
        '</div>',
        className: {
            prevSlide: 'am-pureview-slide-prev',
            nextSlide: 'am-pureview-slide-next'
        },

        selector: {
            close: '[data-am-close="pureview"]',
            total: '.am-pureview-total',
            current: '.am-pureview-current',
            title: '.am-pureview-title'
        }
    };

    PureView.prototype.init = function() {
        var me = this,
            options = me.options,
            $element = me.$element,
            $images = $element.find('img'),
            $pureview = me.$pureview,
            slides = [],
            total = $images.length;

        if (!total) return;

        $images.each(function(i, img) {
            var alt = $(img).attr('alt') || '',
                slide = '<li><img src="' + img.src +'" alt="' + alt + '"/></li>';
            slides.push(slide);
        });

        $pureview.find('.am-pureview-slides').html(slides.join('\n'));

        $('body').append($pureview);

        $pureview.find(options.selector.total).text(total);

        this.$title = $pureview.find(options.selector.title);
        this.$current = $pureview.find(options.selector.current);

        this.$slides = $pureview.find('.am-pureview-slides li');

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
        $pureview.find(options.selector.close).on('click', function(e) {
            e.preventDefault();
            me.close();
        });

        $(document).on('keydown', $.proxy(function(e) {
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
        var $slides = this.$slides,
            activeIndex = $slides.index($slide),
            alt = $slide.children('img').attr('alt');

        alt && this.$title.text(alt);
        this.$current.text(activeIndex + 1);
        $slides.removeAttr('class');
        $slide.addClass('am-active');
        $slides.eq(activeIndex - 1).addClass('am-pureview-slide-prev');
        $slides.eq(activeIndex + 1).addClass('am-pureview-slide-next');
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

    PureView.prototype.open = function(index) {
        var active = index || 0;
        this.activate(this.$slides.eq(active));
        this.$pureview.addClass('am-active');
    };

    PureView.prototype.close = function() {
        this.$pureview.removeClass('am-active');
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
