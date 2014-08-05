define(function(require, exports, module) {
   
    var mask = require('zepto.mask.js');

    var $ = window.Zepto;

    $(".am-switch-mode-ysp").click(function(){
        $(this).amMask.init();
        $(this).amMask.show($('.am-switch-mode-m'));
        return false;
    });

    $('body').click(function(){
        if($('.am-mask').css("display") == "block") {
            $(this).amMask.hide($('.am-switch-mode-m'));
        }
    });

    $('.am-switch-mode-close').click(function(){
        if($('.am-mask').css("display") == "block") {
            $(this).amMask.hide($('.am-switch-mode-m'));
        }
    }); 

});
