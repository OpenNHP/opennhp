define(function(require, exports, module) {
    require('nav');
    require('ui.offcanvas');
    var IScroll = require('iscroll.js');

    var $ = window.Zepto;

    var UI = $.AMUI;

    var menuInit = function() {

        //排除掉one主题
        $('.am-menu').not('[data-am-nav]').not('.am-menu-one').each(function() {
            var nav = $(this);
            if (!nav.data('nav')) {
                var obj = new UI.nav(nav, nav.data('am-nav') ? UI.utils.options(nav.data('am-nav')) : {});
            }
        });

        $('.am-menu-one').each(function() {
            var $this = $(this),
                $warp = $('<div class=\'am-menu-one-warp\'></div>'),
                allWidth = 0,
                prevIndex,
                $nav = $this.find('.am-menu-nav');

            if ($('.am-menu-lv2').length) {

                $this.find('li').each(function() {
                    if ($(this).hasClass('am-parent')) {
                        $(this).attr('am-menu-warp', true);
                        $(this).find('.am-menu-lv2').appendTo($warp);
                    }
                });

                $this.after($warp);
            }

            $nav.wrap('<div class=\'am-menu-nav-wrap\' id=\'am-menu\'>');

            $nav.find('li').eq(0).addClass('am-active');

            // 计算出所有 li 宽度
            $nav.children().each(function(i) {
                allWidth += parseInt($(this).width());
            });

            $nav.width(allWidth);

            var menuScroll = new IScroll('#am-menu', {
                scrollX: true,
                scrollY: false
            });

            $nav.children().on('click', function() {
                $(this).addClass('am-active').siblings().removeClass('am-active');

                // 第一次调用，没有prevIndex
                if (prevIndex === undefined) {
                    prevIndex = $(this).index() ? 0 : 1;
                }

                // 判断方向
                var dir = $(this).index() > prevIndex;
                var target = $(this)[ dir ? 'next' : 'prev' ]();

                // 点击的按钮，显示一半
                var offset = target.offset() || $(this).offset();
                var within = $nav.offset(),
                    listOffset;

                console.log($(this).offset().left);

                if (dir ? offset.left + offset.width > $(document).width() : offset.left + offset.width > 0) {
                    console.log('zheng');
                    listOffset = $nav.offset();

                    //menuScroll.scrollTo( offset.left + offset.width, 0, 400);

                } else {
                    console.log('fu')
                }

                prevIndex = $(this).index();

            });

            $this.on('touchmove', function(event) {
                event.preventDefault();
            })
        });

        function isScrollNext() {

        }
    };

    $(function() {
        menuInit();
    });

    exports.init = menuInit;
});
