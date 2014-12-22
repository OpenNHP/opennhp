# Selected
---

`<select>` 元素样式复写插件。

本插件只提供样式重写及基本的功能，如果需要更高级的功能，请参考：

- [Chosen](https://github.com/harvesthq/chosen)
- [Select2](https://github.com/ivaynberg/select2)
- [bootstrap-select](https://github.com/silviomoreto/bootstrap-select)

## 使用示例

### 单选下拉框

`````html
<select data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select data-am-selected>
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
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
<select data-am-selected="{btnWidth: 300, btnSize: 'sm', btnStyle: 'secondary'}">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select data-am-selected="{btnWidth: 300, btnSize: 'sm', btnStyle: 'secondary'}">
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

### 简易搜索

基于 jQuery `：contains` 选择符实现的简易搜索。

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
<select data-am-selected minchecked="2" maxchecked="3">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
`````
```html
<select data-am-selected minchecked="2" maxchecked="3">
  <option value="a">Apple</option>
  <option value="b">Banana</option>
  <option value="o">Orange</option>
  <option value="m">Mango</option>
</select>
```

## 调用方式

### 通过 Data API

给 `<select>` 元素添加 `data-am-selected` 属性（可设置相关参数），如上面的示例所示。

### 通过 JS

通过 `$('select').selected(options)` 启用样式复写。

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
