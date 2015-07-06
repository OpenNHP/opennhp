'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');

var isWeChat = window.navigator.userAgent.indexOf('MicroMessenger') > -1;

/* global wx,alert */

function appendWeChatSDK(callback) {
  var $weChatSDK = $('<script/>', {
    id: 'wechat-sdk'
  });

  $('body').append($weChatSDK);

  $weChatSDK.on('load', function() {
    callback && callback();
  }).attr('src', 'http://res.wx.qq.com/open/js/jweixin-1.0.0.js');
}

function payHandler() {
  var $paymentBtn = $('[data-am-widget="wechatpay"]');

  if (!isWeChat) {
    $paymentBtn.hide();
    return false;
  }

  $paymentBtn.on('click', '.am-wechatpay-btn', function(e) {
    e.preventDefault();
    var options = UI.utils.parseOptions($(this).parent().data('wechatPay'));
    // console.log(options);
    // alert('pay button clicked');
    if (!window.wx) {
      alert('没有微信 JS SDK');
      return;
    }

    wx.checkJsApi({
      jsApiList: ['chooseWXPay'],
      success: function(res) {
        if (res.checkResult.chooseWXPay) {
          wx.chooseWXPay(options);
        } else {
          alert('微信版本不支持支付接口或没有开启！');
        }
      },
      fail: function() {
        alert('调用 checkJsApi 接口时发生错误!');
      }
    });
  });
}

var payInit = payHandler;

// Init on DOM ready
$(payInit);

module.exports = UI.pay = {
  VERSION: '1.0.0',
  init: payInit
};
