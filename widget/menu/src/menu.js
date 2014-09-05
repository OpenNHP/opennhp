define(function(require, exports, module) {
    require('core');
    require('ui.offcanvas');
    require('ui.collapse');
    var IScroll = require('ui.iscroll-lite');

    var $ = window.Zepto;

    var UI = $.AMUI;

    var menuInit = function() {
        var $menus = $('[data-am-widget="menu"]');

        $menus.find('.am-menu-nav .am-parent > a').on('click', function(e) {
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
        $menus.filter('[data-am-menu-collapse]').find('> .am-menu-toggle').on('click', function(e) {
            e.preventDefault();
            var $this = $(this),
                $nav = $this.next('.am-menu-nav');

            $this.toggleClass('am-active');

            $nav.collapse('toggle');
        });

        // OffCanvas menu
        $menus.filter('[data-am-menu-offcanvas]').find('> .am-menu-toggle').on('click', function(e) {
            e.preventDefault();
            var $this = $(this),
                $nav = $this.next('.am-offcanvas');

            $this.toggleClass('am-active');

            $nav.offCanvas('open');
        });

        // Close offCanvas when link clicked
        var autoCloseOffCanvas = '.am-offcanvas[data-dismiss-on="click"]',
            $autoCloseOffCanvas = $(autoCloseOffCanvas);
        $autoCloseOffCanvas.find('a').not('.am-parent>a').on('click', function(e) {
            $(this).parents(autoCloseOffCanvas).offCanvas('close');
        });

        // one theme
        $menus.filter('.am-menu-one').each(function(index) {
            var $this = $(this),
                $wrap = $('<div class="am-menu-nav-sub-wrap"></div>'),
                allWidth = 0,
                prevIndex,
                $nav = $this.find('.am-menu-nav'),
                $navTopItem = $nav.children('li');

            $navTopItem.filter('.am-parent').each(function(index) {
                $(this).attr('data-rel', '#am-menu-sub-' + index);
                $(this).find('.am-menu-sub').attr('id', 'am-menu-sub-' + index).appendTo($wrap);
            });

            $this.append($wrap);

            $nav.wrap('<div class="am-menu-nav-wrap" id="am-menu-' + index + '">');

            // $navTopItem.eq(0).addClass('am-active');

            // 计算出所有 li 宽度
            $navTopItem.each(function(i) {
                allWidth += parseFloat($(this).css('width'));
            });

            $nav.width(allWidth);

            var menuScroll = new IScroll('#am-menu-' + index, {
                eventPassthrough: true,
                scrollX: true,
                scrollY: false,
                preventDefault: false
            });

            $navTopItem.on('click', function() {
                var $clicked = $(this);
                $clicked.addClass('am-active').siblings().removeClass('am-active');

                $wrap.find('.am-menu-sub.am-in').collapse('close');

                if ($clicked.is('.am-parent')) {
                    !$clicked.hasClass('.am-open') && $wrap.find($clicked.attr('data-rel')).collapse('open');
                } else {
                    $clicked.siblings().removeClass('am-open');
                }

                // 第一次调用，没有prevIndex
                if (prevIndex === undefined) {
                    prevIndex = $(this).index() ? 0 : 1;
                }

                // 判断方向
                var dir = $(this).index() > prevIndex;
                var target = $(this)[dir ? 'next' : 'prev']();

                // 点击的按钮，显示一半
                var offset = target.offset() || $(this).offset();
                var within = $this.offset();

                // 父类左边距

                var listOffset,
                    parentLeft = parseInt($this.css('padding-left'));

                if (dir ? offset.left + offset.width > within.left + within.width : offset.left < within.left) {
                    listOffset = $nav.offset();
                    menuScroll.scrollTo(dir ? within.width - offset.left + listOffset.left - offset.width - parentLeft : listOffset.left - offset.left, 0, 400);
                }

                prevIndex = $(this).index();

            });

            $this.on('touchmove', function(event) {
                event.preventDefault();
            });
        });
    };

    $(function() {
        menuInit();
    });

    exports.init = menuInit;
});
