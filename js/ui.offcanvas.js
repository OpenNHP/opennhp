define(function(require, exports, module) {
    require('zepto.outerdemension');
    require('zepto.extend.data');
    require('core');

    var $ = window.Zepto,
        UI = $.AMUI,
        $win = $(window),
        $doc = $(document),
        scrollPos;

    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/offcanvas.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */

    var OffCanvas = function(element, options) {
        this.$element = $(element);
        this.options = options;
        this.events();
    };

    OffCanvas.DEFAULTS = {
        effect: 'overlay' // {push|overlay}, push is too expensive
    };

    OffCanvas.prototype.open = function(relatedElement) {
        var _self = this,
            $element = this.$element,
            openEvent = $.Event('open:offcanvas:amui');
        
        if (!$element.length || $element.hasClass('am-active')) return;

        var effect = this.options.effect,
            $html = $('html'),
            $bar = $element.find('.am-offcanvas-bar').first(),
            dir = $bar.hasClass('am-offcanvas-bar-flip') ? -1 : 1;

        $bar.addClass('am-offcanvas-bar-' + effect);

        scrollPos = {x: window.scrollX, y: window.scrollY};

        $element.addClass('am-active');

        $html.css({'width': '100%', 'height': $win.height()}).addClass('am-offcanvas-page');

        if (!(effect === 'overlay')) {
            $html.css({'margin-left': $bar.outerWidth() * dir}).width(); // .width() - force redraw
        }

        $html.css('margin-top', scrollPos.y * -1);

        UI.utils.debounce(function() {
            $bar.addClass('am-offcanvas-bar-active').width();
        }, 0)();

        $doc.trigger(openEvent);

        $element.off('.offcanvas.amui').on('click.offcanvas.amui swipeRight.offcanvas.amui swipeLeft.offcanvas.amui', function(e) {
            var $target = $(e.target);

            if (!e.type.match(/swipe/)) {
                if ($target.hasClass('am-offcanvas-bar')) return;
                if ($target.parents('.am-offcanvas-bar').first().length) return;
            }

            // https://developer.mozilla.org/zh-CN/docs/DOM/event.stopImmediatePropagation
            e.stopImmediatePropagation();

            _self.close();
        });

        $doc.on('keydown.offcanvas.amui', function(e) {
            if (e.keyCode === 27) { // ESC
                _self.close();
            }
        });
    };

    OffCanvas.prototype.close = function(relatedElement) {
        var $html = $('html'),
            $element = this.$element,
            $bar = $element.find('.am-offcanvas-bar').first();

        if (!$element.length || !$element.hasClass('am-active')) return;

        $element.trigger('close:offcanvas:amui');

        if (UI.support.transition) {
            $html.one(UI.support.transition.end, function() {
                $html.removeClass('am-offcanvas-page').css({'width': '', 'height': '', 'margin-top': ''});
                $element.removeClass('am-active');
                window.scrollTo(scrollPos.x, scrollPos.y);
            }).css('margin-left', '');

            UI.utils.debounce(function() {
                $bar.removeClass('am-offcanvas-bar-active');
            }, 0)();
        } else {
            $html.removeClass('am-offcanvas-page').attr('style', '');
            $element.removeClass('am-active');
            $bar.removeClass('am-offcanvas-bar-active');
            window.scrollTo(scrollPos.x, scrollPos.y);
        }

        $element.off('.offcanvas.amui');
    };

    OffCanvas.prototype.events = function() {
        $doc.on('click.offcanvas.amui', '[data-am-dismiss="offcanvas"]',
            $.proxy(function(e) {
                e.preventDefault();
                this.close();
            }, this));

        return this;
    };

    UI.offcanvas = OffCanvas;

    function Plugin(option, relatedElement) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.offcanvas'),
                options = $.extend({}, OffCanvas.DEFAULTS, typeof option == 'object' && option);

            if (!data) {
                $this.data('am.offcanvas', (data = new OffCanvas(this, options)));
                data.open(relatedElement);
            }

            if (typeof option == 'string') {
                data[option] && data[option](relatedElement);
            }
        });
    }

    $.fn.offCanvas = Plugin;

    // Init code
    $doc.on('click.offcanvas.amui', '[data-am-offcanvas]', function(e) {
        e.preventDefault();
        var $this = $(this),
            options = UI.utils.parseOptions($this.attr('data-am-offcanvas')),
            $target = $(options.target || (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
            option = $target.data('am.offcanvas') ? 'open' : options;

        Plugin.call($target, option, this);
    });

    module.exports = OffCanvas;
});

// TODO: 优化动画效果
// http://dbushell.github.io/Responsive-Off-Canvas-Menu/step4.html