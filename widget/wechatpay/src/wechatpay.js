'use strict';

var $ = require('jquery');
require('./core');

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
  if (!window.wx) {
    return;
  }

  var $pay = $('[data-am-widget="wechatpay"]');
  var $payBtn = $pay.find('.am-wechatpay-btn');
  var options = $.AMUI.utils.parseOptions($payBtn.data('wechatPay'));
  console.log(options);

  $payBtn.on('click', function() {
    wx.checkJsApi({
      jsApiList: ['chooseWXPay'],
      success: function(res) {
        wx.chooseWXPay({
          timestamp: options.timestamp,  // 支付签名时间戳
          nonceStr: options.nonceStr,  // 支付签名随机串
          package: options.package,  // 统一支付接口返回的prepay_id参数值
          signType: options.signType, // 加密类型
          paySign: options.paySign // 支付签名
        });
      },
      fail: function() {
        alert('你微信当前版本不支持支付功能!');
      }
    });
  });
}

var payInit = function() {
  $('.am-wechatpay').length && payHandler();
};

$(document).on('ready', payInit);

module.exports = $.AMUI.pay = {
  VERSION: '1.0.0',
  init: payInit
};
