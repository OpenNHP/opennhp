define(function (require, exports, module) {

    require('zepto.extend.touch');
    require('core');
    var $ = window.Zepto;
    var tabsInit = function () {
        $('.am-tabs').each(function () {
            amTabs($(this));
        });
    };

    function amTabs(parent) {
        var $tabsContent = parent.find('.am-tabs-bd-content'),
            $tabsDiv = $tabsContent.children(),
            oneWidth,
            iNow = 0,
            disX,
            disY,
            downY,
            downX,
            $tabLi = parent.find('.am-tabs-hd').children();

        //设置tabsdiv宽度
        $tabsContent.width($tabsContent.parent().width() * $tabsDiv.length);
        $tabsDiv.width($tabsContent.parent().width());
        oneWidth = $tabsDiv.width();

        $(window).on('resize', function () {
            $tabsContent.width($tabsContent.parent().width() * $tabsDiv.length);
            $tabsDiv.width($tabsContent.parent().width());
            oneWidth = $tabsDiv.width();
        });

        /*$tabsContent.on("touchstart MSPointerDown pointerdown", function(ev){
         ev.preventDefault();
         var oTarget = ev.targetTouches[0];
         disX = oTarget.clientX - $tabsContent.offset().left;
         disY = oTarget.clientY - $tabsContent.offset().top;
         downX = oTarget.clientX;
         downY = oTarget.clientY;
         $( $tabsContent ).on("touchmove MSPointerMove pointermove", fnMove);
         $( $tabsContent ).on("touchend MSPointerUp pointerup", fnUp);
         });*/

        $tabsContent.swipeRight(function () {

            iNow--;

            if (iNow < 0) {
                iNow = 0;
            }

            $tabsContent.animate({
                'left': -iNow * oneWidth
            });

            $tabLi.removeClass('am-tabs-hd-active');
            $tabLi.eq(iNow).addClass('am-tabs-hd-active');

        });

        $tabsContent.swipeLeft(function () {

            iNow++;

            if (iNow > $tabsDiv.length - 1) {
                iNow = $tabsDiv.length - 1;
            }

            $tabsContent.animate({
                'left': -iNow * oneWidth
            });

            $tabLi.removeClass('am-tabs-hd-active');
            $tabLi.eq(iNow).addClass('am-tabs-hd-active');

        });

        $tabLi.on('click', function () {
            iNow = $(this).index();
            $tabLi.removeClass('am-tabs-hd-active');
            $tabLi.eq(iNow).addClass('am-tabs-hd-active');

            $tabsContent.animate({
                'left': -iNow * oneWidth
            });

        });

        /*
         function fnUp(ev){
         var oTarget = ev.changedTouches[0];
         if( oTarget.clientX - downX < -70 ){
         //('←');
         if( iNow == $tabsDiv.length-1 ){
         //iNow = 0;
         }
         else{
         iNow++;
         }
         $tabsContent.animate({
         "left": -iNow * oneWidth
         })
         }else if(oTarget.clientX - downX > 70){
         //('→');
         if( iNow == 0 ){
         //iNow = 0;
         }
         else{
         iNow--;
         }
         $tabsContent.animate({
         "left": -iNow * oneWidth
         });

         }else{
         $tabsContent.animate({
         "left": -iNow * oneWidth
         });
         }

         $tabLi.removeClass( "am-tabs-hd-active" );
         $tabLi.eq( iNow ).addClass( "am-tabs-hd-active" );

         $( $tabsContent ).off( "touchend MSPointerUp pointerup", fnUp );
         $( $tabsContent ).off( "touchmove MSPointerMove pointermove", fnMove );

         }
         */

        /*function fnMove(ev){
         var oTarget = ev.targetTouches[0];
         ev.stopPropagation();

         if(Math.abs( oTarget.clientX - downX)){
         ev.preventDefault();
         $tabsContent.css( "left", oTarget.clientX - disX );
         }

         if(Math.abs( oTarget.clientY - downY) > 10){
         $('body').scrollTop( $('body').scrollTop() - (oTarget.clientY - downY)/10 );
         }

         }*/
    }

    $(function () {
        tabsInit();
    });

    exports.init = tabsInit;
});
