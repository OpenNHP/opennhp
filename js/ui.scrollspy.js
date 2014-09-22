define(function(require, exports, module) {

    'use strict';

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;


    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/scrollspy.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */

    var ScrollSpy = function(element, options) {

        if (!UI.support.animation) return;

        this.options = $.extend({}, ScrollSpy.DEFAULTS, options);
        this.$element = $(element);

        var checkViewRAF = function() {
            UI.utils.rAF.call(window, $.proxy(this.checkView, this));
        }.bind(this);

        this.$window = $(window).on('scroll.scrollspy.amui', checkViewRAF)
            .on('resize.scrollspy.amui orientationchange.scrollspy.amui', UI.utils.debounce(checkViewRAF, 50));

        this.timer = this.inViewState = this.initInView = null;

        checkViewRAF();
    };

    ScrollSpy.DEFAULTS = {
        animation: 'fade',
        className: {
            inView: 'am-scrollspy-inview',
            init: 'am-scrollspy-init'
        },
        repeat: true,
        delay: 0,
        topOffset: 0,
        leftOffset: 0
    };

    ScrollSpy.prototype.checkView = function () {

        var $element = this.$element,
            options = this.options,
            inView = UI.utils.isInView($element, options),
            animation = (options.animation) ? ' am-animation-' + options.animation : '';

        if(inView && !this.inViewState) {

            if(this.timer) clearTimeout(this.timer);

            if(!this.initInView) {
                $element.addClass(options.className.init);
                this.offset = $element.offset();
                this.initInView = true;

                $element.trigger('init:scrollspy:amui');
            }

            this.timer = setTimeout(function(){
                if(inView) {
                    $element.addClass(options.className.inView + animation).width();
                }
            }, options.delay);

            this.inViewState = true;
            $element.trigger('inview:scrollspy:amui');
        }

        if (!inView && this.inViewState && options.repeat) {
            $element.removeClass(options.className.inView + animation);

            this.inViewState = false;

            $element.trigger('outview:scrollspy:amui');
        }
    };

    ScrollSpy.prototype.check = function() {
        UI.utils.rAF.call(window, $.proxy(this.checkView, this));
    };

    UI.scrollspy = ScrollSpy;


    // Sticky Plugin

    function Plugin(option) {
        return this.each(function () {
            var $this = $(this),
                data = $this.data('am.scrollspy'),
                options = typeof option == 'object' && option;

            if (!data) $this.data('am.scrollspy', (data = new ScrollSpy(this, options)));
            if (typeof option == 'string') data[option]();
        });
    }

    $.fn.scrollspy = Plugin;


    // Init code

    $(function() {
        $('[data-am-scrollspy]').each(function () {
            var $this = $(this),
                options = UI.utils.options($this.attr('data-am-scrollspy'));

            $this.scrollspy(options);
        });
    });


    module.exports = ScrollSpy;
});
