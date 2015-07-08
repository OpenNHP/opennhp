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

A style rewrite of `<select>` element.

This plugin only provide style rewrite and basic functions. If you need more advanced funcitons, see these for more details:

- [Amaze UI Styled Chosen](https://github.com/amazeui/chosen)
- [Chosen](https://github.com/harvesthq/chosen)
- [Select2](https://github.com/ivaynberg/select2)
- [bootstrap-select](https://github.com/silviomoreto/bootstrap-select)

## Example

### Dropdown Single-selection

`````html
<form action="">
  <select name="test" data-am-selected>
    <option value="a">Apple</option>
    <option value="b" selected>Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
    <option value="d" disabled>Disabled</option>
  </select>
</form>
`````
```html
<select data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
  <option value="d" disabled>Disabled</option>
</select>
```

### Dropdown Multi-select

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

### Dropdown Multi-select with Default Value

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

### Grouped Dropdown Multi-select

`````html
<select multiple data-am-selected>
  <optgroup label="Fruits">
    <option value="a">Apple</option>
    <option value="b">Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
  </optgroup>
  <optgroup label="Equipments">
    <option value="phone">iPhone</option>
    <option value="im">iMac</option>
    <option value="mbp">Macbook Pro</option>
  </optgroup>
</select>
`````
```html
<select multiple data-am-selected>
  <optgroup label="Fruits">
    <option value="a">Apple</option>
    <option value="b">Banana</option>
    <option value="o">Orange</option>
    <option value="m">Mango</option>
  </optgroup>
  <optgroup label="Equipments">
    <option value="phone">iPhone</option>
    <option value="im">iMac</option>
    <option value="mbp">Macbook Pro</option>
  </optgroup>
</select>
```

### Size and Color

- `btnWidth`: Button width. Number or percentage. E.g. `btnWidth: '50%'`
- `btnSize`: Button size. `[xl|lg|sm|xs]` (More details in [Button](/css/button?_ver=2.x)).
- `btnStyle`: Button style. `[primary|secondary|success|warning|danger]`

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

### Select with Limited Height

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

### Dropup

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

### Simple Search

Simple search based on jQuery `:contains` selector.

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

### Validation

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

### Control Selected with JS

<span class="am-badge am-badge-danger">New in v2.3!</span>

JS can help inserting, removing and selecting options in `<select>`. When modifying the `<select>`, dropdown menu need to be rerendered.

- Browsers that support [MutationObserver](http://caniuse.com/#search=MutationObserver) will rerender autometically;
- Otherwise, `changed.selected.amui` event need to be manuelly triggered.

**Attention:**

```js
// MutationObserver can observe any use of `attr()`.
$('select').find('option').eq(1).attr('selected', true);

// Following operations won't be observed by MutationObserver
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

<button type="button" data-selected="add" class="am-btn am-btn-primary">Insert Option</button>
<button type="button" data-selected="toggle" class="am-btn am-btn-secondary">Troggle Orange</button>
<button type="button" data-selected="disable" class="am-btn am-btn-danger">Troggle Mango</button>

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
        $selected.append('<option value="o' + i +'">Dynamically inserted option ' + i + '</option>');
        i++;
      }

      if (action === 'toggle') {
        $o.attr('selected', !$o.get(0).selected);
      }

      if (action === 'disable') {
        $m[0].disabled = !$m[0].disabled;
      }

      // When using browsers that don't support MutationObserver, `changed.selected.amui` event need to be manuelly triggered.
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

<button type="button" data-selected="add" class="am-btn am-btn-primary">Insert Option</button>
<button type="button" data-selected="toggle" class="am-btn am-btn-secondary">Troggle Orange</button>
<button type="button" data-selected="disable" class="am-btn am-btn-danger">Troggle Mango</button>

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
      $selected.append('<option value="o' + i +'">Dynamically inserted option ' + i + '</option>');
      i++;
    }

    if (action === 'toggle') {
      $o.attr('selected', !$o.get(0).selected);
    }

    if (action === 'disable') {
      $m[0].disabled = !$m[0].disabled;
    }

    // When using browsers that don't support MutationObserver, `changed.selected.amui` event need to be manuelly triggered.
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

## Usage

### Using Data API

Add `data-am-selected` to `<select>` element as shown above.

### Using JS

Enable style rewrite using `$('select').selected(options)`.

**If the project also contains [jQuery Form](https://github.com/malsup/form/), `$.fn.selected` will cause naming confliction. Please use `$('select').selectIt(options)` instead.**

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

#### Options

- `btnWidth: null`: Button width. Default value is `200px`
- `btnSize: null`: Button size. Optional values include `xl|sm|lg|xl`
- `btnStyle: 'default'`: Button style. Optional values include `primary|secondary|success|warning|danger`
- `maxHeight: null`: The maximum height.
- `dropUp: 0`: Dropup if true. Default value is `0` (`false`)

`````html
      <form data-am-validator>
        <select id="city" data-am-selected="{btnWidth: '200px', btnSize: 'lg', btnStyle: ''}">
          <option value="a">Apple</option>
          <option value="b">Banana</option>
          <option value="o">Orange</option>
          <option value="m">Mango</option>
          <option value="phone">iPhone</option>
          <option value="im">iMac</option>
          <option value="mbp">Macbook Pro</option>
        </select>

      </form>
`````
