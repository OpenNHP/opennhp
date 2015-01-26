var $ = require('jquery');
require('./core');

function addPayApi(callback) {
  var $payApi = $('<script/>', {
    id: 'am-pay-api'
  });

  $('body').append($payApi);

  $payApi.on('load', function() {
    var script = document.createElement('script');
    script.textContent = '(' + callback.toString() + ')();';
    $('body')[0].appendChild(script);
  }).attr('src', 'http://res.wx.qq.com/open/js/jweixin-1.0.0.js');
}

function addPay() {
  var pay = $('[data-am-widget="wechatpay"]');
  var payBtn = pay.find('.am-wechatpay-btn');
  var options = $.AMUI.utils.parseOptions(payBtn.data('wechatPay'));

  wx.ready(function() {
    payBtn.on('click', function() {
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
  });
}

var payInit = function() {
  $('.am-wechatpay').length && addPayApi(addPay);
};

$(document).on('ready', payInit);

module.exports = $.AMUI.pay = {
  VERSION: '1.0.0',
  init: payInit
};
