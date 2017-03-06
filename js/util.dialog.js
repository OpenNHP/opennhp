/**
 * Created by along on 15/8/12.
 */
var $ = require('jquery');
var UI = require('./core');

var dialog = dialog || {};
dialog.alert = function(opt){
    opt = opt || {};
    opt.title = opt.title || '提示';
    opt.content = opt.content || '提示内容';
    opt.onConfirm = opt.onConfirm || function(){};
    var html = [];
    html.push('<div class="am-modal am-modal-alert " tabindex="-1" id="my-alert">');
    html.push('<div class="am-modal-dialog radius">');
    html.push('<div class="am-modal-hd">'+opt.title+'</div>');
    html.push('<div class="am-modal-bd">'+opt.content+'</div>');
    html.push('<div class="am-modal-footer"><span class="am-modal-btn">确定</span></div>');
    html.push('</div>');
    html.push('</div>');
    return $(html.join('')).appendTo('body').modal().on('closed.modal.amui', function() {
        $(this).remove();
        opt.onConfirm();
    });
}

dialog.confirm = function(opt){
    opt = opt || {};
    opt.title = opt.title || '提示';
    opt.content = opt.content || '提示内容';
    opt.onConfirm = opt.onConfirm || function(){};
    opt.onCancel = opt.onCancel || function(){};

    var html = [];
    html.push('<div class="am-modal am-modal-confirm" tabindex="-1">');
    html.push('<div class="am-modal-dialog">');
    html.push('<div class="am-modal-hd">'+opt.title+'</div>');
    html.push('<div class="am-modal-bd">'+opt.content+'</div>');
    html.push('<div class="am-modal-footer">');
    html.push('<span class="am-modal-btn" data-am-modal-cancel>取消</span>');
    html.push('<span class="am-modal-btn" data-am-modal-confirm>确定</span>');
    html.push('</div>');
    html.push('</div>');
    html.push('</div>');

    return $(html.join('')).appendTo('body').modal({
        onConfirm: function(options) {
            opt.onConfirm();
        },
        onCancel: function() {
            opt.onCancel();
        }
    }).on('closed.modal.amui', function() {
        $(this).remove();
    });
}

dialog.loading = function(opt){
    opt = opt || {};
    opt.title = opt.title || '正在载入...';

    var html = [];
    html.push('<div class="am-modal am-modal-loading am-modal-no-btn" tabindex="-1" id="my-modal-loading">');
    html.push('<div class="am-modal-dialog">');
    html.push('<div class="am-modal-hd">'+opt.title+'</div>');
    html.push('<div class="am-modal-bd">');
    html.push('<span class="am-icon-spinner am-icon-spin"></span>');
    html.push('</div>');
    html.push('</div>');
    html.push('</div>');

    return $(html.join('')).appendTo('body').modal().on('closed.modal.amui', function() {
        $(this).remove();
    });
}

dialog.actions = function(opt){
    opt = opt || {};
    opt.title = opt.title || '您想整咋样?';
    opt.items = opt.items || [];
    opt.onSelected = opt.onSelected || function(){
        $acions.close();
    };
    var html = [];
    html.push('<div class="am-modal-actions">');
    html.push('<div class="am-modal-actions-group">');
    html.push('<ul class="am-list">');
    html.push('<li class="am-modal-actions-header">'+opt.title+'</li>');
    opt.items.forEach(function(item,index){
        html.push('<li index="'+index+'">'+item.content+'</li>');
    });
    html.push('</ul>');
    html.push('</div>');
    html.push('<div class="am-modal-actions-group">');
    html.push('<button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button>');
    html.push('</div>');
    html.push('</div>');

    var $acions = $(html.join('')).appendTo('body');
    $acions.find('.am-list>li').bind('click',function(e){
        opt.onSelected($(this).attr('index'),this);
    });
    return {
        show:function(){
            $acions.modal('open');
        },
        close:function(){
            $acions.modal('close');
        }
    };
}

dialog.popup = function(opt){
    opt = opt || {};
    opt.title = opt.title || '标题';
    opt.content = opt.content || '正文';
    opt.onClose = opt.onClose || function(){};
    var html = [];
    html.push('<div class="am-popup">');
    html.push('<div class="am-popup-inner">');
    html.push('<div class="am-popup-hd">');
    html.push('<h4 class="am-popup-title">'+opt.title+'</h4>');
    html.push('<span data-am-modal-close  class="am-close">&times;</span>');
    html.push('</div>');
    html.push('<div class="am-popup-bd">'+opt.content+'</div>');
    html.push('</div> ');
    html.push('</div>');
    return $(html.join('')).appendTo('body').modal().on('closed.modal.amui', function() {
        $(this).remove();
        opt.onClose();
    });
}

module.exports = UI.dialog = dialog;
