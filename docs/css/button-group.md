# Button-group
---

组合 Button 元素。

## 基本使用

把一系列要使用的 `.am-btn` 按钮放入 `.am-btn-group` 。

`````html
<div class="am-btn-group">
    <button type="button" class="am-btn am-btn-default">左手</button>
    <button type="button" class="am-btn am-btn-default">猪手</button>
    <button type="button" class="am-btn am-btn-default">右手</button>
</div>

&nbsp;

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-radius">左手</button>
  <button type="button" class="am-btn am-btn-primary am-radius">猪手</button>
  <button type="button" class="am-btn am-btn-primary am-radius">右手</button>
</div>

&nbsp;

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-success am-round">左手</button>
  <button type="button" class="am-btn am-btn-success am-round">猪手</button>
  <button type="button" class="am-btn am-btn-success am-round">右手</button>
</div>
`````

```html
<div class="am-btn-group">
 <button type="button" class="am-btn am-btn-default">左手</button>
  ...
</div>

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-radius">左手</button>
  ...
</div>

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-round">左手</button>
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
        <button type="button" class="am-btn am-btn-default">左手 lg</button>
        <button type="button" class="am-btn am-btn-default">猪手 lg</button>
        <button type="button" class="am-btn am-btn-default">右手 lg</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group">
        <button type="button" class="am-btn am-btn-default">左手默认</button>
        <button type="button" class="am-btn am-btn-default">猪手默认</button>
        <button type="button" class="am-btn am-btn-default">右手默认</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-sm">
        <button type="button" class="am-btn am-btn-default">左手 sm</button>
        <button type="button" class="am-btn am-btn-default">猪手 sm</button>
        <button type="button" class="am-btn am-btn-default">右手 sm</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-xs">
        <button type="button" class="am-btn am-btn-default">左手 xs</button>
        <button type="button" class="am-btn am-btn-default">猪手 xs</button>
        <button type="button" class="am-btn am-btn-default">右手 xs</button>
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
    <button type="button" class="am-btn am-btn-default">劳资是个按钮</button>
    <button type="button" class="am-btn am-btn-default">劳资是个按钮</button>
    <button type="button" class="am-btn am-btn-default">劳资是个按钮</button>
    <button type="button" class="am-btn am-btn-default">劳资是个按钮</button>
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

~~**注意：** 只适用 `<a>` 元素，`<button>` 不能应用 `display: table-cell` 样式~~。

**使用 `flexbox` 实现，只兼容 IE 10+ 及其他现代浏览器**。

`````html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">左手</a>
  <a class="am-btn am-btn-default" role="button">猪手</a>
  <a class="am-btn am-btn-default" role="button">右手</a>
</div>
`````
```html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">左手</a>
  ...
</div>
```

## 结合下拉组件使用

下面的演示需要结合 [Dropdown](/javascript/dropdown?_ver=2.x) 使用。

### 按钮下拉菜单

`````html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">下拉按钮</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
        <li class="am-dropdown-header">标题</li>
        <li><a href="#">快乐的方式不只一种</a></li>
        <li class="am-active"><a href="#">最荣幸是</a></li>
        <li><a href="#">谁都是造物者的光荣</a></li>
        <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
        <li class="am-divider"></li>
        <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
    </ul>
  </div>
</div>
<script>
$(function() {
  $('[data-am-dropdown]').on('open:dropdown:amui', function () {
    console.log('open event triggered');
  });
});
</script>
`````

```html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">下拉按钮</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">标题</li>
      <li><a href="#">快乐的方式不只一种</a></li>
      <li class="am-active"><a href="#">最荣幸是</a></li>
      <li><a href="#">谁都是造物者的光荣</a></li>
      <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
      <li class="am-divider"></li>
      <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
    </ul>
  </div>
</div>
```

### 按钮上拉菜单

`````html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">上拉按钮</button>
  <div class="am-dropdown am-dropdown-up" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-up"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">标题</li>
      <li><a href="#">快乐的方式不只一种</a></li>
      <li class="am-active"><a href="#">最荣幸是</a></li>
      <li><a href="#">谁都是造物者的光荣</a></li>
      <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
      <li class="am-divider"></li>
      <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
    </ul>
  </div>
</div>
`````

```html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">上拉按钮</button>
  <div class="am-dropdown am-dropdown-up" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-up"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">标题</li>
      <li><a href="#">快乐的方式不只一种</a></li>
      <li class="am-active"><a href="#">最荣幸是</a></li>
      <li><a href="#">谁都是造物者的光荣</a></li>
      <li class="am-disabled"><a href="#">就站在光明的角落</a></li>
      <li class="am-divider"></li>
      <li><a href="#">天空海阔 要做最坚强的泡沫</a></li>
    </ul>
  </div>
</div>
```
