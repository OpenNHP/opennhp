'use strict';

var $ = require('jquery');
var UI = require('./core');
// require('./ui.dropdown');

// Make jQuery :contains Case-Insensitive
$.expr[':'].containsNC = function(elem, i, match, array) {
  return (elem.textContent || elem.innerText || '').toLowerCase().
      indexOf((match[3] || '').toLowerCase()) >= 0;
};

/**
 * Selected
 * @desc HTML select replacer
 * @via https://github.com/silviomoreto/bootstrap-select
 * @license https://github.com/silviomoreto/bootstrap-select/blob/master/LICENSE
 * @param element
 * @param options
 * @constructor
 */

var Selected = function(element, options) {
  this.$element = $(element);
  this.options = $.extend({}, Selected.DEFAULTS, {
    placeholder: element.getAttribute('placeholder') ||
    Selected.DEFAULTS.placeholder
  }, options);
  this.$originalOptions = this.$element.find('option');
  this.multiple = element.multiple;
  this.$selector = null;
  this.initialized = false;
  this.init();
};

Selected.DEFAULTS = {
  btnWidth: null,
  btnSize: null,
  btnStyle: 'default',
  dropUp: 0,
  maxHeight: null,
  maxChecked: null,
  placeholder: '点击选择...',
  selectedClass: 'am-checked',
  disabledClass: 'am-disabled',
  searchBox: false,
  tpl: '<div class="am-selected am-dropdown ' +
  '<%= dropUp ? \'am-dropdown-up\': \'\' %>" id="<%= id %>" data-am-dropdown>' +
  '  <button type="button" class="am-selected-btn am-btn am-dropdown-toggle">' +
  '    <span class="am-selected-status am-fl"></span>' +
  '    <i class="am-selected-icon am-icon-caret-' +
  '<%= dropUp ? \'up\' : \'down\' %>"></i>' +
  '  </button>' +
  '  <div class="am-selected-content am-dropdown-content">' +
  '    <h2 class="am-selected-header">' +
  '<span class="am-icon-chevron-left">返回</span></h2>' +
  '   <% if (searchBox) { %>' +
  '   <div class="am-selected-search">' +
  '     <input autocomplete="off" class="am-form-field am-input-sm" />' +
  '   </div>' +
  '   <% } %>' +
  '    <ul class="am-selected-list">' +
  '      <% for (var i = 0; i < options.length; i++) { %>' +
  '       <% var option = options[i] %>' +
  '       <% if (option.header) { %>' +
  '  <li data-group="<%= option.group %>" class="am-selected-list-header">' +
  '       <%= option.text %></li>' +
  '       <% } else { %>' +
  '       <li class="<%= option.classNames%>" ' +
  '         data-index="<%= option.index %>" ' +
  '         data-group="<%= option.group || 0 %>" ' +
  '         data-value="<%= option.value %>" >' +
  '         <span class="am-selected-text"><%= option.text %></span>' +
  '         <i class="am-icon-check"></i></li>' +
  '      <% } %>' +
  '      <% } %>' +
  '    </ul>' +
  '    <div class="am-selected-hint"></div>' +
  '  </div>' +
  '</div>',
  listTpl:   '<% for (var i = 0; i < options.length; i++) { %>' +
  '       <% var option = options[i] %>' +
  '       <% if (option.header) { %>' +
  '  <li data-group="<%= option.group %>" class="am-selected-list-header">' +
  '       <%= option.text %></li>' +
  '       <% } else { %>' +
  '       <li class="<%= option.classNames %>" ' +
  '         data-index="<%= option.index %>" ' +
  '         data-group="<%= option.group || 0 %>" ' +
  '         data-value="<%= option.value %>" >' +
  '         <span class="am-selected-text"><%= option.text %></span>' +
  '         <i class="am-icon-check"></i></li>' +
  '      <% } %>' +
  '      <% } %>'
};

Selected.prototype.init = function() {
  var _this = this;
  var $element = this.$element;
  var options = this.options;

  $element.hide();

  var data = {
    id: UI.utils.generateGUID('am-selected'),
    multiple: this.multiple,
    options: [],
    searchBox: options.searchBox,
    dropUp: options.dropUp,
    placeholder: options.placeholder
  };

  this.$selector = $(UI.template(this.options.tpl, data));
  // set select button styles
  this.$selector.css({width: this.options.btnWidth});

  this.$list = this.$selector.find('.am-selected-list');
  this.$searchField = this.$selector.find('.am-selected-search input');
  this.$hint = this.$selector.find('.am-selected-hint');

  var $selectorBtn = this.$selector.find('.am-selected-btn');
  var btnClassNames = [];

  options.btnSize && btnClassNames.push('am-btn-' + options.btnSize);
  options.btnStyle && btnClassNames.push('am-btn-' + options.btnStyle);
  $selectorBtn.addClass(btnClassNames.join(' '));

  this.$selector.dropdown({
    justify: $selectorBtn
  });

  // disable Selected instance if <selected> is disabled
  // should call .disable() after Dropdown initialed
  if ($element[0].disabled) {
    this.disable();
  }

  // set list height
  if (options.maxHeight) {
    this.$selector.find('.am-selected-list').css({
      'max-height': options.maxHeight,
      'overflow-y': 'scroll'
    });
  }

  // set hint text
  var hint = [];
  var min = $element.attr('minchecked');
  var max = $element.attr('maxchecked') || options.maxChecked;

  this.maxChecked = max || Infinity;

  if ($element[0].required) {
    hint.push('必选');
  }

  if (min || max) {
    min && hint.push('至少选择 ' + min + ' 项');
    max && hint.push('至多选择 ' + max + ' 项');
  }

  this.$hint.text(hint.join('，'));

  // render dropdown list
  this.renderOptions();

  // append $selector after <select>
  this.$element.after(this.$selector);
  this.dropdown = this.$selector.data('amui.dropdown');
  this.$status = this.$selector.find('.am-selected-status');

  // #try to fixes #476
  setTimeout(function() {
    _this.syncData();
    _this.initialized = true;
  }, 0);

  this.bindEvents();
};

Selected.prototype.renderOptions = function() {
  var $element = this.$element;
  var options = this.options;
  var optionItems = [];
  var $optgroup = $element.find('optgroup');
  this.$originalOptions = this.$element.find('option');

  // 单选框使用 JS 禁用已经选择的 option 以后，
  // 浏览器会重新选定第一个 option，但有一定延迟，致使 JS 获取 value 时返回 null
  if (!this.multiple && ($element.val() === null)) {
    this.$originalOptions.length &&
    (this.$originalOptions.get(0).selected = true);
  }

  function pushOption(index, item, group) {
    if (item.value === '') {
      // skip to next iteration
      // @see http://stackoverflow.com/questions/481601/how-to-skip-to-next-iteration-in-jquery-each-util
      return true;
    }

    var classNames = '';
    item.disabled && (classNames += options.disabledClass);
    !item.disabled && item.selected && (classNames += options.selectedClass);

    optionItems.push({
      group: group,
      index: index,
      classNames: classNames,
      text: item.text,
      value: item.value
    });
  }

  // select with option groups
  if ($optgroup.length) {
    $optgroup.each(function(i) {
      // push group name
      optionItems.push({
        header: true,
        group: i + 1,
        text: this.label
      });

      $optgroup.eq(i).find('option').each(function(index, item) {
        pushOption(index, item, i);
      });
    });
  } else {
    // without option groups
    this.$originalOptions.each(function(index, item) {
      pushOption(index, item, null);
    });
  }

  this.$list.html(UI.template(options.listTpl, {options: optionItems}));
  this.$shadowOptions = this.$list.find('> li').
    not('.am-selected-list-header');
};

Selected.prototype.setChecked = function(item) {
  var options = this.options;
  var $item = $(item);
  var isChecked = $item.hasClass(options.selectedClass);

  if (this.multiple) {
    // multiple
    var checkedLength = this.$list.find('.' + options.selectedClass).length;

    if (!isChecked && this.maxChecked <= checkedLength) {
      this.$element.trigger('checkedOverflow.selected.amui', {
        selected: this
      });

      return false;
    }
  } else {
    // close dropdown whether item is checked or not
    // @see #860
    this.dropdown.close();

    if (isChecked) {
      return false;
    }

    this.$shadowOptions.not($item).removeClass(options.selectedClass);
  }

  $item.toggleClass(options.selectedClass);
  this.syncData(item);
};

/**
 * syncData
 *
 * @description if `item` set, only sync `item` related option
 * @param {Object} [item]
 */
Selected.prototype.syncData = function(item) {
  var _this = this;
  var options = this.options;
  var status = [];
  var $checked = $([]);

  this.$shadowOptions.filter('.' + options.selectedClass).each(function() {
    var $this = $(this);
    status.push($this.find('.am-selected-text').text());

    if (!item) {
      $checked = $checked.add(_this.$originalOptions
        .filter('[value="' + $this.data('value') + '"]')
        .prop('selected', true));
    }
  });

  if (item) {
    var $item = $(item);
    this.$originalOptions
      .filter('[value="' + $item.data('value') + '"]')
      .prop('selected', $item.hasClass(options.selectedClass));
  } else {
    this.$originalOptions.not($checked).prop('selected', false);
  }

  // nothing selected
  if (!this.$element.val()) {
    status = [options.placeholder];
  }

  this.$status.text(status.join(', '));

  // Do not trigger change event on initializing
  this.initialized && this.$element.trigger('change');
};

Selected.prototype.bindEvents = function() {
  var _this = this;
  var header = 'am-selected-list-header';
  var handleKeyup = UI.utils.debounce(function(e) {
    _this.$shadowOptions.not('.' + header).hide().
     filter(':containsNC("' + e.target.value + '")').show();
  }, 100);

  this.$list.on('click', '> li', function(e) {
    var $this = $(this);
    !$this.hasClass(_this.options.disabledClass) &&
      !$this.hasClass(header) && _this.setChecked(this);
  });

  // simple search with jQuery :contains
  this.$searchField.on('keyup.selected.amui', handleKeyup);

  // empty search keywords
  this.$selector.on('closed.dropdown.amui', function() {
    _this.$searchField.val('');
    _this.$shadowOptions.css({display: ''});
  });

  // work with Validator
  // @since 2.5
  this.$element.on('validated.field.validator.amui', function(e) {
    if (e.validity) {
      var valid = e.validity.valid;
      var errorClassName = 'am-invalid';

      _this.$selector[(!valid ? 'add' : 'remove') + 'Class'](errorClassName);
    }
  });

  // observe DOM
  if (UI.support.mutationobserver) {
    this.observer = new UI.support.mutationobserver(function() {
      _this.$element.trigger('changed.selected.amui');
    });

    this.observer.observe(this.$element[0], {
      childList: true,
      subtree: true,
      characterData: true
    });
  }

  // custom event
  this.$element.on('changed.selected.amui', function() {
    _this.renderOptions();
    _this.syncData();
  });
};

// @since: 2.5
Selected.prototype.select = function(item) {
  var $item;

  if (typeof item === 'number') {
    $item = this.$list.find('> li').not('.am-selected-list-header').eq(item);
  } else if (typeof item === 'string') {
    $item = this.$list.find(item);
  } else {
    $item = $(item);
  }

  $item.trigger('click');
};

// @since: 2.5
Selected.prototype.enable = function() {
  this.$element.prop('disable', false);
  this.$selector.dropdown('enable');
};

// @since: 2.5
Selected.prototype.disable = function() {
  this.$element.prop('disable', true);
  this.$selector.dropdown('disable');
};

Selected.prototype.destroy = function() {
  this.$element.removeData('amui.selected').show();
  this.$selector.remove();
};

UI.plugin('selected', Selected);

// Conflict with jQuery form
// https://github.com/malsup/form/blob/6bf24a5f6d8be65f4e5491863180c09356d9dadd/jquery.form.js#L1240-L1258
// https://github.com/allmobilize/amazeui/issues/379
// @deprecated: $.fn.selected = $.fn.selectIt = Plugin;

// New way to resolve conflict:
// @see https://github.com/amazeui/amazeui/issues/781#issuecomment-158873541

UI.ready(function(context) {
  $('[data-am-selected]', context).selected();
});

module.exports = Selected;
