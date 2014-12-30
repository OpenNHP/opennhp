'use strict';

var $ = require('jquery');
var UI = require('./core');

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
  this.options = $.extend({}, Selected.DEFAULTS, options);
  this.$originalOptions = this.$element.find('option');
  this.multiple = element.multiple;
  this.$selector = null;
  this.init();
};

Selected.DEFAULTS = {
  btnWidth: null,
  btnSize: null,
  btnStyle: 'default',
  maxHeight: null,
  noSelectedText: '点击选择...',
  selectedClass: 'am-checked',
  searchBox: false,
  tpl: '<div class="am-selected am-dropdown" id="<%= id %>" data-am-dropdown>' +
  '  <button class="am-selected-btn am-btn am-dropdown-toggle">' +
  '    <span class="am-selected-status am-fl"></span>' +
  '    <i class="am-selected-icon am-icon-caret-down"></i>' +
  '  </button>' +
  '  <div class="am-selected-content am-dropdown-content">' +
  '    <h2 class="am-selected-header"><span class="am-icon-chevron-left">返回</span></h2>' +
  '   <% if (searchBox) { %>' +
  '   <div class="am-selected-search">' +
  '     <input type="text" autocomplete="off" class="am-form-field" />' +
  '   </div>' +
  '   <% } %>' +
  '    <ul class="am-selected-list">' +
  '      <% for (var i = 0; i < options.length; i++) { %>' +
  '       <% var option = options[i] %>' +
  '       <% if (option.header) { %>' +
  '       <li data-group="<%= option.group %>" class="am-selected-list-header">' +
  '       <%= option.text %></li>' +
  '       <% } else { %>' +
  '       <li class=<%= option.active %> ' +
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
  '</div>'
};

Selected.prototype.init = function() {
  var $element = this.$element;
  var options = this.options;
  var optionItems = [];
  var $optgroup = $element.find('optgroup');

  $element.hide();

  function pushOption(index, item, group) {
    optionItems.push({
      group: group,
      index: index,
      active: item.selected ? options.selectedClass : '',
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
    $element.find('option').each(function(index, item) {
      pushOption(index, item, null);
    });
  }

  var data = {
    id: UI.utils.generateGUID('am-selected'),
    multiple: $element.get(0).multiple,
    options: optionItems,
    searchBox: options.searchBox
  };

  this.$selector = $(this.render(data));
  this.$searchField = this.$selector.find('.am-selected-search input');
  this.$hint = this.$selector.find('.am-selected-hint');

  // set select button styles
  var $selectorBtn = this.$selector.find('.am-selected-btn').
    css({width: this.options.btnWidth});
  var btnClassNames = [];

  options.btnSize && btnClassNames.push('am-btn-' + options.btnSize);
  options.btnStyle && btnClassNames.push('am-btn-' + options.btnStyle);
  $selectorBtn.addClass(btnClassNames.join(' '));

  this.$selector.dropdown({
    justify: $selectorBtn
  });

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
  var max = $element.attr('maxchecked');

  if ($element[0].required) {
    hint.push('必选');
  }

  if (min || max) {
    min && hint.push('至少选择 ' + min + ' 项');
    max && hint.push('至多选择 ' + max + ' 项');
  }

  this.$hint.text(hint.join('，'));

  // append $selector after <select>
  this.$element.after(this.$selector);
  this.dropdown = this.$selector.data('amui.dropdown');
  this.$status = this.$selector.find('.am-selected-status');

  this.getShadowOptions();
  this.syncData();
  this.bindEvents();
};

Selected.prototype.render = function(data) {
  return UI.template(this.options.tpl, data);
};

Selected.prototype.getShadowOptions = function() {
  this.$shadowOptions = this.$selector.find('.am-selected-list li').
    not('.am-selected-list-header');
};

Selected.prototype.setChecked = function(item) {
  var options = this.options;
  var $item = $(item);
  var isChecked = $item.hasClass(options.selectedClass);
  if (!this.multiple) {
    if (!isChecked) {
      this.dropdown.close();
      this.$shadowOptions.not($item).removeClass(options.selectedClass);
    } else {
      return;
    }
  }

  $item.toggleClass(options.selectedClass);

  this.syncData(item);
};

/**
 * syncData
 * @desc if `item` set, only sync `item` related option
 * @param {Object} item
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
      $checked = $checked.add(_this.$originalOptions.
        filter('[value="' + $this.data('value') + '"]').
        prop('selected', true));
    }
  });

  if (item) {
    var $item = $(item);
    this.$originalOptions.filter('[value="' + $item.data('value') + '"]').
      prop('selected', $item.hasClass(options.selectedClass));
  } else {
    this.$originalOptions.not($checked).prop('selected', false);
  }

  // nothing selected
  if (!this.$element.val()) {
    status.push(options.noSelectedText);
  }

  this.$status.text(status.join(', '));
  this.$element.trigger('change');
};

Selected.prototype.bindEvents = function() {
  var _this = this;
  var handleKeyup = UI.utils.debounce(function(e) {
    _this.$shadowOptions.not('.am-selcted-list-header').hide().
     filter(':containsNC("' + e.target.value + '")').show();
  }, 100);

  this.$shadowOptions.on('click', function(e) {
    _this.setChecked(this);
  });

  // simple search with jQuery :contains
  this.$searchField.on('keyup.selected.amui', handleKeyup);

  // empty search keywords
  this.$selector.on('closed.dropdown.amui', function() {
    _this.$searchField.val('');
    _this.$shadowOptions.css({display: ''});
  });
};

Selected.prototype.destroy = function() {
  this.$element.removeData('amui.selected').show();
  this.$selector.remove();
};

function Plugin(option) {
  return this.each(function() {
    var $this = $(this);
    var data = $this.data('amui.selected');
    var options = $.extend({}, UI.utils.parseOptions($this.data('amSelected')),
      typeof option === 'object' && option);

    if (!data && option === 'destroy') {
      return;
    }

    if (!data) {
      $this.data('amui.selected', (data = new Selected(this, options)));
    }

    if (typeof option == 'string') {
      data[option] && data[option]();
    }
  });
}

$.fn.selected = Plugin;

UI.ready(function(context) {
  $('[data-am-selected]', context).selected();
});

$.AMUI.selected = Selected;

module.exports = Selected;
