---
id: selected
title: 下拉选框样式增强
titleEn: Slected
prev: javascript/datepicker.html
next: javascript/ucheck.html
source: js/ui.selected.js
doc: docs/javascript/selected.md
---

# Selected
---

`<select>` 元素样式复写插件。

本插件只提供样式重写及基本的功能，如果需要更高级的功能，请参考：

- [Amaze UI Styled Chosen](https://github.com/amazeui/chosen)
- [Chosen](https://github.com/harvesthq/chosen)
- [Select2](https://github.com/ivaynberg/select2)
- [bootstrap-select](https://github.com/silviomoreto/bootstrap-select)

## 使用示例

### 单选下拉框

`````html
<form action="">
  <select name="test" data-am-selected>
    <option value="a">Apple</option>
    <option value="b" selected>Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
    <option value="d" disabled>禁用鸟</option>
  </select>
</form>
`````
```html
<select data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="d" disabled>禁用鸟</option>
</select>
```

### 多选下拉框

`````html
<select multiple data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select multiple data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
```

### 多选下拉框 - 有默认选中项

`````html
<select multiple data-am-selected>
  <option value="a">Apple</option>
  <option value="b" selected>Banana</option>
  <option value="o">Orange</option>
  <option value="m" selected>Mango</option>
</select>
`````
```html
<select multiple data-am-selected>
  <option value="a">Apple</option>
  <option value="b" selected>Banana</option>
  <option value="o">Orange</option>
  <option value="m" selected>Mango</option>
</select>
```

### 分组多选下拉框

`````html
<select multiple data-am-selected>
  <optgroup label="水果">
    <option value="a">Apple</option>
    <option value="b">Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
  </optgroup>
  <optgroup label="装备">
    <option value="phone">iPhone</option>
    <option value="im">iMac</option>
    <option value="mbp">Macbook Pro</option>
  </optgroup>
</select>
`````
```html
<select multiple data-am-selected>
  <optgroup label="水果">
    <option value="a">Apple</option>
    <option value="b">Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
  </optgroup>
  <optgroup label="装备">
    <option value="phone">iPhone</option>
    <option value="im">iMac</option>
    <option value="mbp">Macbook Pro</option>
  </optgroup>
</select>
```

### 按钮尺寸及颜色

- `btnWidth`: 按钮宽度，数字或者百分比，`btnWidth: '50%'`
- `btnSize`: 按钮尺寸，`[xl|lg|sm|xs]` （[参见 Button](/css/button?_ver=2.x)）
- `btnStyle`: 按钮风格，`[primary|secondary|success|warning|danger]`

`````html
<select data-am-selected="{btnWidth: '40%', btnSize: 'sm', btnStyle: 'secondary'}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select data-am-selected="{btnWidth: '40%', btnSize: 'sm', btnStyle: 'secondary'}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
```

### 限制列表高度

`````html
<select data-am-selected="{maxHeight: 100}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
`````
```html
<select data-am-selected="{maxHeight: 100}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
```

### 上拉选择

`````html
<select data-am-selected="{dropUp: 1}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
`````
```html
<select data-am-selected="{dropUp: 1}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
```

### 简易搜索

基于 jQuery `:contains` 选择符实现的简易搜索。

`````html
<select data-am-selected="{searchBox: 1}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
`````
```html
<select data-am-selected="{searchBox: 1}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="phone">iPhone</option>
  <option value="im">iMac</option>
  <option value="mbp">Macbook Pro</option>
</select>
```

### 选项验证提示

`````html
<select multiple data-am-selected minchecked="2" maxchecked="3">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select multiple data-am-selected minchecked="2" maxchecked="3">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
```

### JS 操作 select

<span class="am-badge am-badge-danger">v2.3 新增！</span>

使用 JS 操作 `<select>`（如添加选项、禁用选项、选中选项等），需要重新渲染下拉菜单。

- 支持 [MutationObserver](http://caniuse.com/#search=MutationObserver) 的浏览器会自动触发重新渲染；
- 其他浏览器需要手动触发 `changed.selected.amui` 事件。

**需要注意的是**：

```js
// 使用 `attr()` 可以被 MutationObserver 观察到
$('select').find('option').eq(1).attr('selected', true);

// 以下操作不会被 MutationObserver 观察到
$('select').val('aa');
$('select').find('option').eq(1).prop('selected', true);
$('select').find('option')(1).selected = true;
```

`````html
<select id="js-selected" data-am-selected>
  <option value="a">Apple</option>
  <option value="b" selected>Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>

<hr/>

<button type="button" data-selected="add" class="am-btn am-btn-primary">添加选项</button>
<button type="button" data-selected="toggle" class="am-btn am-btn-secondary">交替 Orange 选中状态</button>
<button type="button" data-selected="disable" class="am-btn am-btn-danger">交替 Mango 禁用状态</button>

<hr/>
<div id="js-selected-info"></div>
<script>
  $(function() {
    var $selected = $('#js-selected');
    var $o = $selected.find('option[value="o"]');
    var $m = $selected.find('option[value="m"]');
    var i = 0;

    $('[data-selected]').on('click', function() {
      var action = $(this).data('selected');

      if (action === 'add') {
        $selected.append('<option value="o' + i +'">动态插入的选项 ' + i + '</option>');
        i++;
      }

      if (action === 'toggle') {
        $o.attr('selected', !$o.get(0).selected);
      }

      if (action === 'disable') {
        $m[0].disabled = !$m[0].disabled;
      }

      // 不支持 MutationObserver 的浏览器使用 JS 操作 select 以后需要手动触发 `changed.selected.amui` 事件
      if (!$.AMUI.support.mutationobserver) {
        $selected.trigger('changed.selected.amui');
      }
    });

    $selected.on('change', function() {
      $('#js-selected-info').html([
        '选中项：<strong class="am-text-danger">',
        [$(this).find('option').eq(this.selectedIndex).text()],
        '</strong> 值：<strong class="am-text-warning">',
        $(this).val(),
        '</strong>'
      ].join(''));
    });
  });
</script>
`````

```html
<select id="js-selected" data-am-selected>
  <option value="a">Apple</option>
  <option value="b" selected>Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>

<hr/>

<button type="button" data-selected="add" class="am-btn am-btn-primary">添加选项</button>
<button type="button" data-selected="toggle" class="am-btn am-btn-secondary">交替 Orange 选中状态</button>
<button type="button" data-selected="disable" class="am-btn am-btn-danger">交替 Mango 禁用状态</button>

<hr/>
<div id="js-selected-info"></div>
```

```js
$(function() {
  var $selected = $('#js-selected');
  var $o = $selected.find('option[value="o"]');
  var $m = $selected.find('option[value="m"]');
  var i = 0;

  $('[data-selected]').on('click', function() {
    var action = $(this).data('selected');

    if (action === 'add') {
      $selected.append('<option value="o' + i +'">动态插入的选项 ' + i + '</option>');
      i++;
    }

    if (action === 'toggle') {
      $o.attr('selected', !$o.get(0).selected);
    }

    if (action === 'disable') {
      $m[0].disabled = !$m[0].disabled;
    }

    // 不支持 MutationObserver 的浏览器使用 JS 操作 select 以后需要手动触发 `changed.selected.amui` 事件
    if (!$.AMUI.support.mutationobserver) {
      $selected.trigger('changed.selected.amui');
    }
  });

  $selected.on('change', function() {
    $('#js-selected-info').html([
      '选中项：<strong class="am-text-danger">',
      [$(this).find('option').eq(this.selectedIndex).text()],
      '</strong> 值：<strong class="am-text-warning">',
      $(this).val(),
      '</strong>'
    ].join(''));
  });
});
```

## 调用方式

### 通过 Data API

给 `<select>` 元素添加 `data-am-selected` 属性（可设置相关参数），如上面的示例所示。

### 通过 JS

通过 `$('select').selected(options)` 启用样式复写。

**如果项目中同时使用了 [jQuery Form](https://github.com/malsup/form/)，`$.fn.selected` 有命名冲突，请使用 `$('select').selectIt(options)` 替代。**

```javascript
$(function() {
  // 使用默认参数
  $('select').selected();

  // 设置参数
  $('select').selected({
    btnWidth: '300px',
    btnSize: 'sm',
    btnStyle: 'primary',
    maxHeight: '100px'
  });
});
```

#### 参数说明

- `btnWidth: null`: 按钮宽度，默认为 `200px`
- `btnSize: null`: 按钮尺寸，可选值为 `xl|sm|lg|xl`
- `btnStyle: 'default'`: 按钮样式，可选值为 `primary|secondary|success|warning|danger`
- `maxHeight: null`: 列表最大高度
- `dropUp: 0`: 是否为上拉，默认为 `0` (`false`)
