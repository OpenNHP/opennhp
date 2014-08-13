define(function(require, exports, module) {
    require('ui.offcanvas');
    require('ui.collapse');

    var $ = window.Zepto;

    var UI = $.AMUI;

    var menuInit = function() {

        var $menus = $('[data-am-widget="menu"]');

        $menus.find('.am-menu-nav .am-parent > a').on('click', function(e) {
            e.preventDefault();
            var $clicked = $(this),
                $parent= $clicked.parent(),
                $subMenu= $clicked.next('.am-menu-sub');
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


        // has one class
        if ($('.am-menu').hasClass('am-menu-one')) {
            var $this, $next, $width, iNow, aNum, $menuLv2, $menuLv3;


            $this = $('.am-menu-one');
            $next = $('<a>').attr({class: 'am-menu-next', href: 'javascript:;'});
            $width = 0;
            iNow = -1;
            aNum = [];
            $menuLv2 = $('.am-menu-lv2');
            $menuLv3 = $('.am-menu-lv3');

            $this.children('li').find('a').eq(0).attr('href', 'javascript:;').addClass('am-menu-prev am-menu-disabled');
            $this.find('li').eq(0).append($next);
            $this.find('.am-menu-lv2').wrap('<div class=\'am-menu-wrap\'></div>');

            $menuLv2.children('li').children('a').each(function() {
                $(this).parent().width($(this).width());
                $width += $(this).width();
                aNum.push($width);
            });

            $menuLv2.find('.am-parent').each(function() {
                var $firstA = $(this).find('a'),
                    $li = $('<li class=\'am-menu-item-more\'><a class=\'am-menu-item-close\' href=\'javascript:;\'>×</a><a class=\'am-menu-item-into\' href=' + $firstA.attr('href') + '>进入' + $firstA.html() + '</a></li>');

                $firstA.attr('href', 'javascript:;');
                $(this).find('.am-menu-lv3').append($li);
            });
            $menuLv2.width($width);
            $menuLv3.width($('.am-menu-lv1').width() - 20);// 减去Menu 左右padding
            // FIXME: ide border
            //$menuLv3.width($('.am-menu-lv1').width() - 22);// 减去Menu 左右padding

            $('.am-menu-wrap .am-parent').children('a').on('click', function() {
                if ($(this).hasClass('active')) {
                    $(this).removeClass('active');
                    $(this).siblings('.am-menu-lv3').animate({opacity: 0}, 'fast', 'linear', function() {
                        $(this).css('display', 'none');
                    });
                } else {
                    offAll();
                    $(this).addClass('active');
                    $(this).siblings('.am-menu-lv3').css('display', 'block').animate({
                        left: -$(this).offset().left + 10,
                        opacity: 1
                    }, 'fast', 'linear');
                    // FIXME: ide offset不一致，多加17px
                    //$(this).siblings('.am-menu-lv3').css('display', 'block').animate({left: -$(this).offset().left + 27, opacity: 1}, 'fast', 'linear');
                }
            });

            $('.am-menu-next').on('click', function() {
                offAll();
                if (-aNum[iNow] + parseInt($menuLv2.css('left')) < (-$menuLv2.width() + $menuLv2.parent().width() - $next.width() * 2)) {
                    $menuLv2.animate({left: -$menuLv2.width() + $menuLv2.parent().width() - $next.width() * 2}, 'fast', 'linear');
                    $(this).addClass('am-menu-disabled');
                } else {
                    iNow++;
                    $menuLv2.animate({left: -aNum[iNow]}, 'fast', 'linear');
                    $('.am-menu-prev').removeClass('am-menu-disabled');
                }
            });

            $('.am-menu-item-close').on('click', function() {
                offAll();
            });

            $('.am-menu-prev').on('click', function() {
                offAll();
                if (iNow <= -1) {
                    $menuLv2.animate({left: 0}, 'fast', 'linear');
                    $(this).addClass('am-menu-disabled');
                } else {
                    iNow--;
                    $menuLv2.animate({left: -aNum[iNow]}, 'fast', 'linear');
                    $('.am-menu-next').removeClass('am-menu-disabled');
                }
            });

            drag($menuLv2);
        }

        /*
         *  offAll menu children active
         */
        function offAll() {
            $('.am-menu-wrap .am-parent')
                .children('a')
                .removeClass('active')
                .siblings('.am-menu-lv3')
                .animate({opacity: 0}, 'fast', 'linear')
                .css('display', 'none');
        }

        /*
         *  drag menu children
         *  @obj Zepto object
         */
        function drag(obj) {
            var disX,
                downX,
                nOffsetLeft = 0;
            obj.on('touchstart MSPointerDown pointerdown', function(ev) {
                offAll();
                ev.preventDefault();
                var oTarget = ev.targetTouches[0];
                disX = oTarget.clientX - $(this).offset().left;
                downX = oTarget.clientX;

                $(document).on('touchmove MSPointerMove pointermove', fnMove);
                $(document).on('touchend MSPointerUp pointerup', fnUp);

            });

            function fnUp(ev) {
                $.each(aNum, function(index, item) {
                    nOffsetLeft += -aNum[index];
                    if (parseInt(obj.css('left')) >= nOffsetLeft) {
                        iNow = index;
                        return false;
                    }
                });

                nOffsetLeft = 0;
                $(document).off('touchend MSPointerUp pointerup', fnUp);
                $(document).off('touchmove MSPointerMove pointermove', fnMove);
            }

            function fnMove(ev) {
                ev.preventDefault();
                var oTarget = ev.targetTouches[0];
                var nLeft = oTarget.clientX - disX;
                // ->
                if (nLeft > 0) {
                    nLeft = 0;
                }
                // <-
                if (nLeft < -obj.width() + obj.parent().width() - $next.width() * 2) {
                    nLeft = -obj.width() + obj.parent().width() - $next.width() * 2;
                }
                obj.css('left', nLeft);
            }
        }
    };

    $(function() {
        menuInit();
    });

    exports.init = menuInit;
});
