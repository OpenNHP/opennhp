'use strict';

var $ = require('jquery');
require('./core');

function listNewsInit() {
  $('.am-list-news-one').each(function() {
    amListNewsMore($(this));
  });
}

function amListNewsMore($element) {
  var $amList = $element.find('.am-list');

  var $listMore = '<a class="am-list-news-more am-btn am-btn-default" ' +
    'href="javascript:;">更多 &gt;&gt;</a>';

  if ($amList.children().length > 6) {

    $amList.children().each(function(index) {
      if (index > 5) {
        $(this).hide();
      }
    });

    $element.find('.am-list-news-more').remove();
    $element.append($listMore);
  }

  $element.find('.am-list-news-more').on('click', function() {
    $amList.children().show();
    $(this).hide();
  });
}

$(function() {
  listNewsInit();
});

module.exports = $.AMUI.listNews = {
  VERSION: '3.0.0',
  init: listNewsInit
};
