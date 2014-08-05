define(function(require, exports, module) {

    'use strict';

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;


    // Sticky Class

    var Sticky = function(element, options) {

        this.options = $.extend({}, Sticky.DEFAULTS, options);
        this.$element = $(element);

        this.$window = $(window)
            .on('scroll.sticky.amui', UI.utils.debounce($.proxy(this.checkPosition, this), 50))
            .on('click.sticky.amui', UI.utils.debounce($.proxy(this.checkPosition, this), 1));

        this.original = {
                offsetTop: this.$element.offset().top,
                width: this.$element.width()
            };

        this.sticked = null;

        this.checkPosition();
    };

    Sticky.DEFAULTS = {
        top: 0,
        cls: 'am-sticky'
    };

    Sticky.prototype.checkPosition = function () {

        if (!this.$element.is(':visible')) return;

        var scrollHeight = $(document).height(),
            scrollTop = this.$window.scrollTop(),
            options = this.options,
            offsetTop = options.top,
            $element = this.$element,
            animation = (options.animation) ? ' am-animation-' + options.animation : "";

        this.sticked = (scrollTop > this.original.offsetTop) ? 'sticky' : false;

        if (this.sticked) {
            $element.addClass(options.cls + animation).css({top: offsetTop});
        } else {
            $element.removeClass(options.cls + animation).css({top: ''});
        }
    };

    UI.sticky = Sticky;


    // Sticky Plugin

    function Plugin(option) {
        return this.each(function () {
            var $this   = $(this),
                data    = $this.data('am.sticky'),
                options = typeof option == 'object' && option;

            if (!data) $this.data('am.sticky', (data = new Sticky(this, options)));
            if (typeof option == 'string') data[option]();
        });
    }

    $.fn.sticky = Plugin;


    // Init code

    $(window).on('load', function () {
        $('[data-am-sticky]').each(function () {
            var $this = $(this),
                options = UI.utils.options($this.attr('data-am-sticky'));

            Plugin.call($this, options);
        });
    });


    module.exports = Sticky;
});
