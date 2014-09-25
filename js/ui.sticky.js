define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/addons/sticky.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */

    // Sticky Class

    var Sticky = function(element, options) {
        var me = this;

        this.options = $.extend({}, Sticky.DEFAULTS, options);
        this.$element = $(element);
        this.sticked = null;
        this.inited = null;
        this.$holder = undefined;

        this.$window = $(window).
            on('scroll.sticky.amui', UI.utils.debounce($.proxy(this.checkPosition, this), 10)).
            on('resize.sticky.amui orientationchange.sticky.amui', UI.utils.debounce(function() {
                me.reset(true, function() {
                    me.checkPosition();
                });
            }, 50)).
            on('load.sticky.amui', $.proxy(this.checkPosition, this));

        this.offset = this.$element.offset();

        this.init();
    };

    Sticky.DEFAULTS = {
        top: 0,
        bottom: 0,
        animation: '',
        className: {
            sticky: 'am-sticky',
            resetting: 'am-sticky-resetting',
            stickyBtm: 'am-sticky-bottom',
            animationRev: 'am-animation-reverse'
        }
    };

    Sticky.prototype.init = function() {
        var result = this.check();

        if (!result) return false;

        var $element = this.$element,
            $holder = $('<div class="am-sticky-placeholder"></div>').css({
                'height': $element.css('position') != 'absolute' ? $element.outerHeight() : '',
                'float': $element.css('float') != 'none' ? $element.css('float') : '',
                'margin': $element.css('margin')
            });

        this.$holder = $element.css('margin', 0).wrap($holder).parent();

        this.inited = 1;

        return true;
    };

    Sticky.prototype.reset = function(force, cb) {
        var options = this.options,
            $element = this.$element,
            animation = (options.animation) ? ' am-animation-' + options.animation : "",
            complete = function() {
                $element.css({position: '', top: '', width: '', left: '', margin: 0});
                $element.removeClass([animation, options.className.animationRev, options.className.sticky, options.className.resetting].join(' '));

                this.animating = false;
                this.sticked = false;
                this.offset = $element.offset();
                cb && cb();
            }.bind(this);

        $element.addClass(options.className.resetting);

        if (!force && options.animation && UI.support.animation) {

            this.animating = true;

            $element.removeClass(animation).one(UI.support.animation.end, function() {
                complete();
            }).width(); // force redraw

            $element.addClass(animation + ' ' + options.className.animationRev);
        } else {
            complete();
        }
    };

    Sticky.prototype.check = function() {
        if (!this.$element.is(':visible')) return false;

        var media = this.options.media;

        if (media) {
            switch (typeof(media)) {
                case 'number':
                    if (window.innerWidth < media) {
                        return false;
                    }
                    break;

                case 'string':
                    if (window.matchMedia && !window.matchMedia(media).matches) {
                        return false;
                    }
                    break;
            }
        }

        return true;
    };

    Sticky.prototype.checkPosition = function() {
        if (!this.inited) {
            var initialized = this.init();
            if (!initialized) return;
        }

        var options = this.options,
            scrollHeight = $('body').height(),
            scrollTop = this.$window.scrollTop(),
            offsetTop = options.top,
            offsetBottom = options.bottom,
            $element = this.$element,
            animation = (options.animation) ? ' am-animation-' + options.animation : "",
            className = [options.className.sticky, animation].join(' ');

        if (typeof offsetBottom == 'function') offsetBottom = offsetBottom(this.$element);

        var checkResult = (scrollTop > this.$holder.offset().top);

        if (!this.sticked && checkResult) {
            $element.addClass(className);
        } else if (this.sticked && !checkResult) {
            this.reset();
        }

        this.$holder.height($element.height());

        if (checkResult) {
            $element.css({
                top: offsetTop,
                left: this.$holder.offset().left,
                width: this.offset.width
            });
            
            /*
            if (offsetBottom) {
                // （底部边距 + 元素高度 > 窗口高度） 时定位到底部
                if ((offsetBottom + this.offset.height > $(window).height()) &&
                    (scrollTop + $(window).height() >= scrollHeight - offsetBottom)) {
                    $element.addClass(options.className.stickyBtm).css({top: $(window).height() - offsetBottom - this.offset.height});
                } else {
                    $element.removeClass(options.className.stickyBtm).css({top: offsetTop});
                }
            }
            */
        }

        this.sticked = checkResult;
    };

    UI.sticky = Sticky;


    // Sticky Plugin

    function Plugin(option) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.sticky'),
                options = typeof option == 'object' && option;

            if (!data) $this.data('am.sticky', (data = new Sticky(this, options)));
            if (typeof option == 'string') data[option]();
        });
    }

    $.fn.sticky = Plugin;


    // Init code
    $(window).on('load', function() {
        $('[data-am-sticky]').each(function() {
            var $this = $(this),
                options = UI.utils.options($this.attr('data-am-sticky'));

            Plugin.call($this, options);
        });
    });


    module.exports = Sticky;
});
