define(function(require, exports, module) {

    require('core');
    var Hammer = require('util.hammer');

    var $ = window.Zepto,
        UI = $.AMUI,
        supportTransition = UI.support.transition,
        animation = UI.support.animation;

    /**
     * @via https://github.com/twbs/bootstrap/blob/master/js/tab.js
     * @copyright 2011-2014 Twitter, Inc.
     * @license MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE)
     */

    var Tabs = function(element) {
        this.$element = $(element);

        this.$tabNav = this.$element.find(Tabs.DEFAULTS.selector.nav);
        this.$navs = this.$tabNav.find('a');

        this.$content = this.$element.find(Tabs.DEFAULTS.selector.content);
        this.$tabPanels = this.$content.find(Tabs.DEFAULTS.selector.panel);

        this.transitioning = null;
        
        this.init();
    };

    Tabs.DEFAULTS = {
        selector: {
            nav: '.am-tabs-nav',
            content: '.am-tabs-bd',
            panel: '.am-tab-panel'
        },
        className: {
            active: 'am-active'
        }
    };

    Tabs.prototype.init = function() {
        var me = this;

        this.$navs.on('click.tabs.amui', $.proxy(function(e) {
            e.preventDefault();
            this.open($(e.target));
        }, this));
        
        var hammer = new Hammer(this.$content[0]);

        hammer.get('pan').set({direction: Hammer.DIRECTION_HORIZONTAL, threshold: 30});

        hammer.on('panleft', UI.utils.debounce(function(e) {
                $(e.target).focus();
                var $nav = me.getNextNav($(e.target));
                $nav && me.open($nav);
            }, 100));

        hammer.on('panright', UI.utils.debounce(function(e) {
            var $nav = me.getPrevNav($(e.target));
            $nav && me.open($nav);
        }, 100));
    };

    Tabs.prototype.open = function($nav) {
        if (!$nav || this.transitioning || $nav.parent('li').hasClass('am-active')) return;

        var $tabNav = this.$tabNav,
            $navs = this.$navs,
            $tabContent = this.$content,
            href = $nav.attr('href'),
            regexHash = /^#.+$/,
            $target = regexHash.test(href) && this.$content.find(href) || this.$tabPanels.eq($navs.index($nav));

        var previous = $tabNav.find('.am-active a')[0],
            e = $.Event('open:tabs:amui', {
                relatedTarget: previous
            });

        $nav.trigger(e);

        if (e.isDefaultPrevented()) return;

        // activate Tab nav
        this.activate($nav.closest('li'), $tabNav);

        // activate Tab content
        this.activate($target, $tabContent, function() {
            $nav.trigger({
                type: 'opened:tabs:amui',
                relatedTarget: previous
            })
        })
    };

    Tabs.prototype.activate = function($element, $container, callback) {
        this.transitioning = true;

        var $active = $container.find('> .am-active'),
            transition = callback && supportTransition && !!$active.length;

        $active.removeClass('am-active am-in');

        $element.addClass('am-active');

        if (transition) {
            $element.redraw(); // reflow for transition
            $element.addClass('am-in');
        } else {
            $element.removeClass('am-fade');
        }

        function complete() {
            callback();
            this.transitioning = false;
        }

        transition ? $active.one(supportTransition.end, $.proxy(complete, this)) : $.proxy(complete, this);
    };

    Tabs.prototype.getNextNav = function($panel) {
        var navIndex = this.$tabPanels.index(($panel)),
            rightSpring = 'am-animation-right-spring';

        if (navIndex + 1 >= this.$navs.length) { // last one
            animation && $panel.addClass(rightSpring).on(animation.end, function() {
                $panel.removeClass(rightSpring);
            });
            return null;
        } else {
            return this.$navs.eq(navIndex + 1);
        }
    };

    Tabs.prototype.getPrevNav = function($panel) {
        var navIndex = this.$tabPanels.index(($panel)),
            leftSpring = 'am-animation-left-spring';

        if (navIndex === 0) { // first one
            animation && $panel.addClass(leftSpring).on(animation.end, function() {
                $panel.removeClass(leftSpring);
            });
            return null;
        } else {
            return this.$navs.eq(navIndex - 1);
        }
    };

    // Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this),
                $tabs = $this.is('.am-tabs') && $this || $this.closest('.am-tabs'),
                data = $tabs.data('amui.tabs');

            if (!data) $tabs.data('amui.tabs', (data = new Tabs($tabs[0])));
            if (typeof option == 'string' && $this.is('.am-tabs-nav a')) data[option]($this);
        });
    }

    $.fn.tabs = Plugin;

    // Init code
    $(document).on('ready', function(e) {
        $('[data-am-tabs]').tabs();
    });

    module.exports = Tabs;
});

// TODO: 1. Ajax 支持
