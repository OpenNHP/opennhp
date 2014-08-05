# Button-group
---

组合 Button 元素。

## 基本使用

把一系列要使用的 `.am-btn` 按钮放入 `.am-btn-group` 。

`````html
<div class="am-btn-group">
    <button type="button" class="am-btn am-btn-default">Left</button>
    <button type="button" class="am-btn am-btn-default">Middle</button>
    <button type="button" class="am-btn am-btn-default">Right</button>
</div>

&nbsp;

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary">Left</button>
  <button type="button" class="am-btn am-btn-primary">Middle</button>
  <button type="button" class="am-btn am-btn-primary">Right</button>
</div>
`````

```html
<div class="am-btn-group">
 <button type="button" class="am-btn am-btn-default">Left</button>
  ...
</div>

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary">Left</button>
  ...
</div>
`````

## 按钮工具栏

将 `.am-btn-group` 放进 `.am-btn-toolbar`，实现工具栏效果。


`````html
<div class="am-btn-toolbar">
  <div class="am-btn-group">
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-align-left"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-align-center"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-align-right"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-align-justify"></i></button>
  </div>
  <div class="am-btn-group">
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-font"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-bold"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-italic"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-underline"></i></button>
  </div>
  <div class="am-btn-group">
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-copy"></i></button>
    <button type="button" class="am-btn am-btn-primary"><i class="am-icon-paste"></i></button>
  </div>
</div>
`````

```html
<div class="am-btn-toolbar">
  <div class="am-btn-group">...</div>
  <div class="am-btn-group">...</div>
  <div class="am-btn-group">...</div>
</div>
```

## 按钮组大小

给 `.am-btn-group ` 增加 class `.am-btn-group-lg` 或 `.am-btn-group-sm` 或 `.am-btn-group-xs` 改变按钮大小。


`````html
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-lg">
        <button type="button" class="am-btn am-btn-default">Left</button>
        <button type="button" class="am-btn am-btn-default">Middle</button>
        <button type="button" class="am-btn am-btn-default">Right</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group">
        <button type="button" class="am-btn am-btn-default">Left</button>
        <button type="button" class="am-btn am-btn-default">Middle</button>
        <button type="button" class="am-btn am-btn-default">Right</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-sm">
        <button type="button" class="am-btn am-btn-default">Left</button>
        <button type="button" class="am-btn am-btn-default">Middle</button>
        <button type="button" class="am-btn am-btn-default">Right</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-xs">
        <button type="button" class="am-btn am-btn-default">Left</button>
        <button type="button" class="am-btn am-btn-default">Middle</button>
        <button type="button" class="am-btn am-btn-default">Right</button>
    </div>
</div>
`````

```html
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-lg">...</div>
</div>

<div class="am-btn-toolbar">
    <div class="am-btn-group">...</div>
</div>

<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-sm">...</div>
</div>

<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-xs">...</div>
</div>
</div>
```

## 垂直排列

使用 `.am-btn-group-stacked` 使按钮垂直排列显示。

`````html
<div class="am-btn-group-stacked">
    <button type="button" class="am-btn am-btn-default">Button</button>
    <button type="button" class="am-btn am-btn-default">Button</button>
    <button type="button" class="am-btn am-btn-default">Button</button>
    <button type="button" class="am-btn am-btn-default">Button</button>
</div>
`````
```html
<div class="am-btn-group-stacked">
  <button type="button" class="am-btn am-btn-default">Button</button>
  ...
</div>
```

## 自适应宽度

添加 `.am-btn-group-justify` class 让按钮组里的按钮平均分布，填满容器宽度。

**注意：** 只适用 `<a>` 元素，`<button>` 不能应用 `display: table-cell` 样式。


`````html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">Left</a>
  <a class="am-btn am-btn-default" role="button">Middle</a>
  <a class="am-btn am-btn-default" role="button">Right</a>
</div>
`````
```html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">Left</a>
  ...
</div>
```

## 按钮下拉菜单

`````html
<div class="am-btn-group">
  <button class="am-btn am-btn-primary">Button</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. 一行代码，简单快捷</a></li>
      <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
      <li><a href="#">3. 内容实时同步更新</a></li>
      <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. 专属的一键拨叫</a></li>
    </ul>
  </div>
</div>
<script>
  seajs.use(['ui.dropdown'], function() {
    $(function() {
      $('[data-am-dropdown]').on('open:dropdown:amui', function () {
        console.log('open event triggered');
      });
    });
  });
</script>
`````
  
```html
<div class="am-btn-group">
  <button class="am-btn am-btn-primary">Button</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. 一行代码，简单快捷</a></li>
      <li class="am-active"><a href="#">2. 网址不变且唯一</a></li>
      <li><a href="#">3. 内容实时同步更新</a></li>
      <li class="am-disabled"><a href="#">4. 云端跨平台适配</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. 专属的一键拨叫</a></li>
    </ul>
  </div>
</div>
```