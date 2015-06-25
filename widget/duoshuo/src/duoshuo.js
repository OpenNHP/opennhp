'use strict';

var $ = require('jquery');
var UI = require('../../../js/core');

function duoshuoInit() {
  var $dsThread = $('.ds-thread');
  var dsShortName = $dsThread.parent('[data-am-widget="duoshuo"]').
    attr('data-ds-short-name');
  var dsSrc = (document.location.protocol == 'https:' ? 'https:' : 'http:') +
    '//static.duoshuo.com/embed.js';

  if (!$dsThread.length || !dsShortName) {
    return;
  }

  window.duoshuoQuery = {
    short_name: dsShortName
  };

  // 已经有多说脚本
  if ($('script[src="' + dsSrc + '"]').length) {
    return;
  }

  var $dsJS = $('<script>', {
    async: true,
    type: 'text/javascript',
    src: dsSrc,
    charset: 'utf-8'
  });

  $('body').append($dsJS);
}

$(window).on('load', duoshuoInit);

module.exports = UI.duoshuo = {
  VERSION: '2.0.1',
  init: duoshuoInit
};
