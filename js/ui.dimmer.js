'use strict';

var $ = require('jquery');
var UI = require('./core');
var $doc = $(document);
var transition = UI.support.transition;

var Dimmer = function() {
  this.id = UI.utils.generateGUID('am-dimmer');
  this.$element = $(Dimmer.DEFAULTS.tpl, {
    id: this.id
  });

  this.inited = false;
  this.scrollbarWidth = 0;
  this.$used = $([]);
};

Dimmer.DEFAULTS = {
  tpl: '<div class="am-dimmer" data-am-dimmer></div>'
};

Dimmer.prototype.init = function() {
  if (!this.inited) {
    $(document.body).append(this.$element);
    this.inited = true;
    $doc.trigger('init.dimmer.amui');
    this.$element.on('touchmove.dimmer.amui', function(e) {
      e.preventDefault();
    });
  }

  return this;
};

Dimmer.prototype.open = function(relatedElement) {
  if (!this.inited) {
    this.init();
  }

  var $element = this.$element;

  // 用于多重调用
  if (relatedElement) {
    this.$used = this.$used.add($(relatedElement));
  }

  this.checkScrollbar().setScrollbar();

  $element.show().trigger('open.dimmer.amui');

  transition && $element.off(transition.end);

  setTimeout(function() {
    $element.addClass('am-active');
  }, 0);

  return this;
};

Dimmer.prototype.close = function(relatedElement, force) {
  this.$used = this.$used.not($(relatedElement));

  if (!force && this.$used.length) {
    return this;
  }

  var $element = this.$element;

  $element.removeClass('am-active').trigger('close.dimmer.amui');

  function complete() {
    $element.hide();
    this.resetScrollbar();
  }

  // transition ? $element.one(transition.end, $.proxy(complete, this)) :
  complete.call(this);

  return this;
};

Dimmer.prototype.checkScrollbar = function() {
  this.scrollbarWidth = UI.utils.measureScrollbar();

  return this;
};

Dimmer.prototype.setScrollbar = function() {
  var $body = $(document.body);
  var bodyPaddingRight = parseInt(($body.css('padding-right') || 0), 10);

  if (this.scrollbarWidth) {
    $body.css('padding-right', bodyPaddingRight + this.scrollbarWidth);
  }

  $body.addClass('am-dimmer-active');

  return this;
};

Dimmer.prototype.resetScrollbar = function() {
  $(document.body).css('padding-right', '').removeClass('am-dimmer-active');

  return this;
};

module.exports = UI.dimmer = new Dimmer();
