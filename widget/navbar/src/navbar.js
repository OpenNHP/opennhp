define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto,
        qrcode = require('util.qrcode'),
        modal = require('ui.modal');

    var navbarInit = function() {
        var _parent = $('.am-navbar'),
            parentUl = _parent.find('ul'),
            subLi = _parent.find('li'),
            oneWidth = subLi.width(),
            minWidth = 100, //每个li最小宽度
            _more = null,
            _moreList = null,
            onOff = true,
            onOffCreat = true, // 防止多次的创建
            $body = $('body');

        var $share = $('[data-am-navbar-share]');
        var $qrcode = $('[data-am-navbar-qrcode]');

        var navbarPosition = _parent.css('position');

        if (navbarPosition == 'fixed') {
            $body.addClass('with-fixed-navbar');
        }

        if ($qrcode.length) {

            var qrImg = ($('[data-am-navbar-qrcode]').attr('data-am-navbar-qrcode'));
            var url = window.location.href;
            var qrData = $('<div class="am-modal am-modal-no-btn" id=\'am-navbar-boxqrcode\'>' +
            '<div class=\'am-modal-dialog\' id=\'am-navbar-qrcode-data\'></div>' +
            '</div>');

            $body.append(qrData);

            //判断上传自定义的二维码没有，否则生成二维码
            if (qrImg) {
                $('#am-navbar-qrcode-data').html('<img src="' + qrImg + '"/>');
            } else {
                var qrnode = new qrcode({
                    render: 'canvas',
                    correctLevel: 0,
                    text: url,
                    width: 190,
                    height: 190,
                    background: '#fff',
                    foreground: '#000'
                });
                $('#am-navbar-qrcode-data').html(qrnode);
            }
        }


        //添加share className
        $share.addClass('am-navbar-share');
        $qrcode.addClass('am-navbar-qrcode');

        if ($share.length) {
            //share start
            window._bd_share_config = {
                "common": {
                    "bdSnsKey": {},
                    "bdText": "",
                    "bdMini": "2",
                    "bdMiniList": false,
                    "bdPic": "",
                    "bdStyle": "1",
                    "bdSize": "16"
                }, "share": {"bdSize": 24}
            };

            $body.append($('<script />', {
                src: 'http://bdimg.share.baidu.com/static/api/js/share.js?v=89343201.js?cdnversion=' + ~(-new Date() / 36e5)
            }));

            var shareData = '<div class="bdsharebuttonbox">' +

                '<div class="am-modal-actions am-modal-out" id="am-navbar-share">' +

                '<div class="am-modal-actions-group">' +
                '<ul class="am-list">' +
                '<li class="am-modal-actions-header" data-cmd="more">分享到</li>' +
                '<li><a href="#" class="am-icon-qq" data-cmd="qzone" title="分享到QQ空间">QQ空间</a></li>' +
                '<li><a href="#" class="am-icon-weibo" data-cmd="tsina" title="分享到新浪微博">新浪微博</a></li>' +
                '<li><a href="#" class="am-icon-tencent-weibo" data-cmd="tqq" title="分享到腾讯微博">腾讯微博</a></li>' +
                '<li><a href="#" class="am-icon-renren" data-cmd="renren" title="分享到人人网">人人网</a></li>' +
                '<li><a href="#" class="am-icon-wechat" data-cmd="weixin" title="分享到微信">微信</a></li>' +
                '</ul>' +
                '</div>' +
                '<div class="am-modal-actions-group"><button type="button" class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button></div>' +
                '</div>' +

                '</div>';

            $body.append(shareData);

            $share.on('click', function(event) {
                event.preventDefault();
                $('#am-navbar-share').modal();
            });

            //share end
        }

        if ($qrcode.length) {
            //qrcode start
            $qrcode.on('click', function(event) {
                event.preventDefault();
                $('#am-navbar-boxqrcode').modal();
            });
        }

        //qrcode end
        if (_parent.length) {
            $body.append($('<ul class="am-navbar-actions"></ul>'));
        }
        if (_parent.find('li').length * _parent.find('li').width() > $(window).width()) { //如果li没有完全展示
            //替换父级的class
            displaceClass(_parent.find('li').length, parentUl);
            var nowWidth = _parent.find('li').width();
            if (nowWidth < minWidth) {
                if (onOffCreat) {
                    addMore();
                    onOffCreat = false;
                }
                displaceClass(liLength(), parentUl);
                addMoreLi(liLength());
            }
        }

        _more = $('.am-navbar-more');
        _moreList = $('.am-navbar-actions');

        _parent.on('click', '.am-navbar-more', function() {
            if (onOff) {
                _moreList.css({
                    bottom: _moreList.height(),
                    display: 'block'
                }).animate({
                    bottom: 49
                }, {
                    duration: 'fast',
                    complete: function() {
                        _more.addClass('am-navbar-more-active');
                    }
                });
                onOff = !onOff;
            } else {
                _moreList.animate({
                    bottom: -_moreList.height()
                }, {
                    complete: function() {
                        $(this).css('display', 'none');
                        _more.removeClass('am-navbar-more-active');
                    }
                });
                onOff = !onOff;
            }
        });

        //添加more

        function addMore() {
            parentUl.append($('<li class="am-navbar-item am-navbar-more"><a href="javascript:;"><span class="am-icon-chevron-up"></span>更多</a></li>'));
        }

        //删除more
        function removeMore() {
            parentUl.find('.am-navbar-more').remove();
        }

        //计算合适的长度
        function liLength() {
            return parseInt($(window).width() / minWidth);
        }

        //移出parent下的li,并添加到moreList里面
        function addMoreLi(len) {
            subLi.not('.am-navbar-more').each(function(index) {
                if (index > len - 2) {
                    $(this).appendTo($('.am-navbar-actions'))
                }
            })
        }

        //移出moreList里面的li,并添加到parent下面
        function addParentLi(len) {
            $('.am-navbar-actions').children().first().appendTo(parentUl)
        }

        //替换class
        function displaceClass(num, object) {
            var $className = object.attr('class').replace(/sm-block-grid-\d/, 'sm-block-grid-' + num);
            object.attr('class', $className);
        }
    };

    // DOMContentLoaded
    $(function() {
        navbarInit();
    });

    exports.init = navbarInit;
});
