define(function (require, exports, module) {
    require('ui.offcanvas');
    var IScroll = require('ui.iscroll');

    var $ = window.Zepto;

    var UI = $.AMUI;

    var menuInit = function () {
        var $menus = $('[data-am-widget="menu"]');

        $menus.find('.am-menu-nav .am-parent > a').on('click', function (e) {
            e.preventDefault();
            var $clicked = $(this),
                $parent = $clicked.parent(),
                $subMenu = $clicked.next('.am-menu-sub');
            $parent.toggleClass('am-open');
            $subMenu.collapse('toggle');
            $parent.siblings('.am-parent').removeClass('am-open')
                .children('.am-menu-sub.am-in').collapse('close');
        });

        // Dropdown/slidedown menu
        $menus.filter('[data-am-menu-collapse]').find('> .am-menu-toggle').on('click', function (e) {
            e.preventDefault();
            var $this = $(this),
                $nav = $this.next('.am-menu-nav');

            $this.toggleClass('am-active');

            $nav.collapse('toggle');
        });

        // OffCanvas menu
        $menus.filter('[data-am-menu-offcanvas]').find('> .am-menu-toggle').on('click', function (e) {
            e.preventDefault();
            var $this = $(this),
                $nav = $this.next('.am-offcanvas');

            $this.toggleClass('am-active');

            $nav.offCanvas('open');
        });

        // one theme
        $menus.filter('.am-menu-one').each(function () {
            var $this = $(this),
                $wrap = $('<div class="am-menu-nav-sub-wrap"></div>'),
                allWidth = 0,
                prevIndex,
                $nav = $this.find('.am-menu-nav');

            $this.find('.am-parent').each(function () {
                $(this).find('.am-menu-sub').appendTo($wrap);

            });

            $this.append($wrap);

            $nav.wrap('<div class="am-menu-nav-wrap" id="am-menu">');

            $nav.find('li').eq(0).addClass('am-active');

            // 计算出所有 li 宽度
            $nav.children().each(function (i) {
                allWidth += parseInt($(this).width());
            });

            $nav.width(allWidth);

            var menuScroll = new IScroll('#am-menu', {
                scrollX: true,
                scrollY: false
            });

            $nav.children().on('click', function () {
                var $clicked = $(this);
                $clicked.addClass('am-active').siblings().removeClass('am-active');

                $wrap.find('.am-menu-sub.am-in').collapse('close');

                if ($clicked.is('.am-parent')) {
                    !$clicked.hasClass('.am-open') && $('.am-menu-sub').eq($clicked.index()).collapse('open');
                } else {
                    $clicked.siblings().removeClass('am-open');
                }

                // 第一次调用，没有prevIndex
                if (prevIndex === undefined) {
                    prevIndex = $(this).index() ? 0 : 1;
                }

                // 判断方向
                var dir = $(this).index() > prevIndex;
                var target = $(this)[ dir ? 'next' : 'prev' ]();

                // 点击的按钮，显示一半
                var offset = target.offset() || $(this).offset();
                var within = $nav.offset();

                if (dir ? offset.left + offset.width > $(document).width() : offset.left < 10) {
                    menuScroll.scrollTo(dir ? within.left - offset.width - 10 : within.left - offset.left, 0, 400);
                }

                prevIndex = $(this).index();

            });

            $this.on('touchmove', function (event) {
                event.preventDefault();
            });
        });
    };

    $(function () {
        menuInit();
    });

    exports.init = menuInit;
});
