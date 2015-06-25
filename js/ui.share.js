'use strict';

require('./ui.modal');
var $ = require('jquery');
var UI = require('./core');
var QRCode = require('./util.qrcode');
var doc = document;
var $doc = $(doc);

var Share = function(options) {
  this.options = $.extend({}, Share.DEFAULTS, options || {});
  this.$element = null;
  this.$wechatQr = null;
  this.pics = null;
  this.inited = false;
  this.active = false;
  // this.init();
};

Share.DEFAULTS = {
  sns: ['weibo', 'qq', 'qzone', 'tqq', 'wechat', 'renren'],
  title: '分享到',
  cancel: '取消',
  closeOnShare: true,
  id: UI.utils.generateGUID('am-share'),
  desc: 'Hi，孤夜观天象，发现一个不错的西西，分享一下下 ;-)',
  via: 'Amaze UI',
  tpl: '<div class="am-share am-modal-actions" id="<%= id %>">' +
  '<h3 class="am-share-title"><%= title %></h3>' +
  '<ul class="am-share-sns am-avg-sm-3">' +
  '<% for(var i = 0; i < sns.length; i++) {%>' +
  '<li>' +
  '<a href="<%= sns[i].shareUrl %>" ' +
  'data-am-share-to="<%= sns[i].id %>" >' +
  '<i class="am-icon-<%= sns[i].icon %>"></i>' +
  '<span><%= sns[i].title %></span>' +
  '</a></li>' +
  '<% } %></ul>' +
  '<div class="am-share-footer">' +
  '<button class="am-btn am-btn-default am-btn-block" ' +
  'data-am-share-close><%= cancel %></button></div>' +
  '</div>'
};

Share.SNS = {
  weibo: {
    title: '新浪微博',
    url: 'http://service.weibo.com/share/share.php',
    width: 620,
    height: 450,
    icon: 'weibo'
  },
  // url          链接地址
  // title:”,     分享的文字内容(可选，默认为所在页面的title)
  // appkey:”,    您申请的应用appkey,显示分享来源(可选)
  // pic:”,       分享图片的路径(可选)
  // ralateUid:”, 关联用户的UID，分享微博会@该用户(可选)
  // NOTE: 会自动抓取图片，不用指定 pic

  qq: {
    title: 'QQ 好友',
    url: 'http://connect.qq.com/widget/shareqq/index.html',
    icon: 'qq'
  },
  // url:,
  // title:'',    分享标题(可选)
  // pics:'',     分享图片的路径(可选)
  // summary:'',  分享摘要(可选)
  // site:'',     分享来源 如：腾讯网(可选)
  // desc: ''     发送给用户的消息
  // NOTE: 经过测试，最终发给用户的只有 url 和 desc

  qzone: {
    title: 'QQ 空间',
    url: 'http://sns.qzone.qq.com/cgi-bin/qzshare/cgi_qzshare_onekey',
    icon: 'star'
  },
  // http://sns.qzone.qq.com/cgi-bin/qzshare/cgi_qzshare_onekey?url=xxx&title=xxx&desc=&summary=&site=
  // url:,
  // title:'',    分享标题(可选)
  // desc:'',     默认分享理由(可选)
  // summary:'',  分享摘要(可选)
  // site:'',     分享来源 如：腾讯网(可选)
  // pics:'',     分享图片的路径(可选)，不会自动抓取，多个图片用|分隔

  tqq: {
    title: '腾讯微博',
    url: 'http://v.t.qq.com/share/share.php',
    icon: 'tencent-weibo'
  },
  // url=xx&title=&appkey=801cf76d3cfc44ada52ec13114e84a96
  // url
  // title
  // pic 多个图片用 | 分隔
  // appkey
  // NOTE: 不会自动抓取图片

  wechat: {
    title: '微信',
    url: '[qrcode]',
    icon: 'wechat'
  },
  // 生成一个二维码 供用户扫描
  // 相关接口 https://github.com/zxlie/WeixinApi

  renren: {
    title: '人人网',
    url: 'http://widget.renren.com/dialog/share',
    icon: 'renren'
  },
  // http://widget.renren.com/dialog/share?resourceUrl=www&srcUrl=www&title=ww&description=xxx
  // 550 * 400
  // resourceUrl : '', // 分享的资源Url
  // srcUrl : '',	     // 分享的资源来源Url,
  //                   //   默认为header中的Referer,如果分享失败可以调整此值为resourceUrl试试
  // pic : '',		 // 分享的主题图片，会自动抓取
  // title : '',		 // 分享的标题
  // description : ''	 // 分享的详细描述
  // NOTE: 经过测试，直接使用 url 参数即可

  douban: {
    title: '豆瓣',
    url: 'http://www.douban.com/recommend/',
    icon: 'share-alt'
  },
  // http://www.douban.com/service/sharebutton
  // 450 * 330
  // http://www.douban.com/share/service?bm=1&image=&href=xxx&updated=&name=
  // href 链接
  // name 标题

  /* void (function() {
   var d = document, e = encodeURIComponent,
   s1 = window.getSelection, s2 = d.getSelection,
   s3 = d.selection, s = s1 ? s1()
   : s2 ? s2() : s3 ? s3.createRange().text : '',
   r = 'http://www.douban.com/recommend/?url=&title=&sel=&v=1&r=1'
   })();
   */

  // tsohu: '',
  // http://t.sohu.com/third/post.jsp?url=&title=&content=utf-8&pic=

  // print: '',

  mail: {
    title: '邮件分享',
    url: 'mailto:',
    icon: 'envelope-o'
  },

  sms: {
    title: '短信分享',
    url: 'sms:',
    icon: 'comment'
  }
};

Share.prototype.render = function() {
  var options = this.options;
  var snsData = [];
  var title = encodeURIComponent(doc.title);
  var link = encodeURIComponent(doc.location);
  var msgBody = '?body=' + title + link;

  options.sns.forEach(function(item, i) {
    if (Share.SNS[item]) {
      var tmp = Share.SNS[item];
      var shareUrl;

      tmp.id = item;

      if (item === 'mail') {
        shareUrl = msgBody + '&subject=' + options.desc;
      } else if (item === 'sms') {
        shareUrl = msgBody;
      } else {
        shareUrl = '?url=' + link + '&title=' + title;
      }

      tmp.shareUrl = tmp.url + shareUrl;

      snsData.push(tmp);
    }
  });

  return UI.template(options.tpl, $.extend({}, options, {sns: snsData}));
};

Share.prototype.init = function() {
  if (this.inited) {
    return;
  }

  var me = this;
  var shareItem = '[data-am-share-to]';

  $doc.ready($.proxy(function() {
    $('body').append(this.render()); // append share DOM to body
    this.$element = $('#' + this.options.id);

    this.$element.find('[data-am-share-close]').
      on('click.share.amui', function() {
        me.close();
      });
  }, this));

  $doc.on('click.share.amui', shareItem, $.proxy(function(e) {
    var $clicked = $(e.target);
    var $target = $clicked.is(shareItem) && $clicked ||
      $clicked.parent(shareItem);
    var sns = $target.attr('data-am-share-to');

    if (!(sns === 'mail' || sns === 'sms')) {
      e.preventDefault();
      this.shareTo(sns, this.setData(sns));
    }

    this.close();
  }, this));

  this.inited = true;
};

Share.prototype.open = function() {
  !this.inited && this.init();
  this.$element && this.$element.modal('open');
  this.$element.trigger('open.share.amui');
  this.active = true;
};

Share.prototype.close = function() {
  this.$element && this.$element.modal('close');
  this.$element.trigger('close.share.amui');
  this.active = false;
};

Share.prototype.toggle = function() {
  this.active ? this.close() : this.open();
};

Share.prototype.setData = function(sns) {
  if (!sns) {
    return;
  }

  var shareData = {
    url: doc.location,
    title: doc.title
  };
  var desc = this.options.desc;
  var imgSrc = this.pics || [];
  var qqReg = /^(qzone|qq|tqq)$/;

  if (qqReg.test(sns) && !imgSrc.length) {
    var allImages = doc.images;

    for (var i = 0; i < allImages.length && i < 10; i++) {
      !!allImages[i].src && imgSrc.push(encodeURIComponent(allImages[i].src));
    }

    this.pics = imgSrc; // 缓存图片
  }

  switch (sns) {
    case 'qzone':
      shareData.desc = desc;
      shareData.site = this.options.via;
      shareData.pics = imgSrc.join('|');
      // TODO: 抓取图片多张
      break;
    case 'qq':
      shareData.desc = desc;
      shareData.site = this.options.via;
      shareData.pics = imgSrc[0];
      // 抓取一张图片
      break;
    case 'tqq':
      // 抓取图片多张
      shareData.pic = imgSrc.join('|');
      break;
  }

  return shareData;
};

Share.prototype.shareTo = function(sns, data) {
  var snsInfo = Share.SNS[sns];

  if (!snsInfo) {
    return;
  }

  if (sns === 'wechat' || sns === 'weixin') {
    return this.wechatQr();
  }

  var query = [];
  for (var key in data) {
    if (data[key]) {
      // 避免 encode 图片分隔符 |
      query.push(key.toString() + '=' + ((key === 'pic' || key === 'pics') ?
        data[key] : encodeURIComponent(data[key])));
    }
  }

  window.open(snsInfo.url + '?' + query.join('&'));
};

Share.prototype.wechatQr = function() {
  if (!this.$wechatQr) {
    var qrId = UI.utils.generateGUID('am-share-wechat');
    var $qr = $('<div class="am-modal am-modal-no-btn am-share-wechat-qr">' +
    '<div class="am-modal-dialog"><div class="am-modal-hd">分享到微信 ' +
    '<a href="" class="am-close am-close-spin" ' +
    'data-am-modal-close>&times;</a> </div>' +
    '<div class="am-modal-bd">' +
    '<div class="am-share-wx-qr"></div>' +
    '<div class="am-share-wechat-tip">' +
    '打开微信，点击底部的<em>发现</em>，<br/> ' +
    '使用<em>扫一扫</em>将网页分享至朋友圈</div></div></div></div>');

    $qr.attr('id', qrId);

    var qrNode = new QRCode({
      render: 'canvas',
      correctLevel: 0,
      text: doc.location.href,
      width: 180,
      height: 180,
      background: '#fff',
      foreground: '#000'
    });

    $qr.find('.am-share-wx-qr').html(qrNode);

    $qr.appendTo($('body'));

    this.$wechatQr = $('#' + qrId);
  }

  this.$wechatQr.modal('open');
};

var share = new Share();

$doc.on('click.share.amui.data-api', '[data-am-toggle="share"]', function(e) {
  e.preventDefault();
  share.toggle();
});

module.exports = UI.share = share;
