# Button
---

## 基本使用

### 默认样式

在要应用按钮样式的元素上添加 `.am-btn`，再设置相应的颜色。

- `.am-btn-default` - 默认，灰色按钮
- `.am-btn-primary` - 蓝色按钮
- `.am-btn-secondary` - 浅蓝色按钮
- `.am-btn-success` - 绿色按钮
- `.am-btn-warning` - 橙色按钮
- `.am-btn-danger` - 红色按钮
- `.am-btn-link`

`````html
<button type="button" class="am-btn am-btn-default">默认样式</button>
<button type="button" class="am-btn am-btn-primary">主色按钮</button>
<button type="button" class="am-btn am-btn-secondary">次色按钮</button>
<button type="button" class="am-btn am-btn-success">绿色按钮</button>
<button type="button" class="am-btn am-btn-warning">橙色按钮</button>
<button type="button" class="am-btn am-btn-danger">红色按钮</button>
<a class="am-btn am-btn-link">链接</a>
<a class="am-btn am-btn-danger" href="http://www.bing.com" target="_blank">应用按钮样式的链接</a>
`````

```html
<button type="button" class="am-btn am-btn-default">默认样式</button>
<button type="button" class="am-btn am-btn-primary">主色按钮</button>
<button type="button" class="am-btn am-btn-secondary">次色按钮</button>
<button type="button" class="am-btn am-btn-success">绿色按钮</button>
<button type="button" class="am-btn am-btn-warning">橙色按钮</button>
<button type="button" class="am-btn am-btn-danger">红色按钮</button>
<a class="am-btn am-btn-link">链接</a>
<a class="am-btn am-btn-danger" href="http://www.bing.com" target="_blank">应用按钮样式的链接</a>
```

### 圆角按钮

在默认样式的基础上添加 `.am-radius` class.

`````html
<button type="button" class="am-btn am-btn-default am-radius">默认样式</button>
<button type="button" class="am-btn am-btn-primary am-radius">主色按钮</button>
<button type="button" class="am-btn am-btn-secondary am-radius">次色按钮</button>
<button type="button" class="am-btn am-btn-success am-radius">绿色按钮</button>
<button type="button" class="am-btn am-btn-warning am-radius">橙色按钮</button>
<button type="button" class="am-btn am-btn-danger am-radius">红色按钮</button>
`````
```html
<button type="button" class="am-btn am-btn-default am-radius">默认样式</button>
```

### 椭圆形按钮

在默认样式的基础上添加 `.am-round` class.

`````html
<button type="button" class="am-btn am-btn-default am-round">默认样式</button>
<button type="button" class="am-btn am-btn-primary am-round">主色按钮</button>
<button type="button" class="am-btn am-btn-secondary am-round">次色按钮</button>
<button type="button" class="am-btn am-btn-success am-round">绿色按钮</button>
<button type="button" class="am-btn am-btn-warning am-round">橙色按钮</button>
<button type="button" class="am-btn am-btn-danger am-round">红色按钮</button>
`````
```html
<button type="button" class="am-btn am-btn-default am-round">默认样式</button>
```

## 按钮状态

### 激活状态

在按钮上添加 `.am-active` class。

`````html
<button type="button" class="am-btn am-btn-primary am-active">激活状态</button>
<button type="button" class="am-btn am-btn-default am-active">激活状态</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">链接按钮激活状态</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">链接按钮激活状态</a>
`````
```html
<button type="button" class="am-btn am-btn-primary am-active">激活状态</button>
<button type="button" class="am-btn am-btn-default am-active">激活状态</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">链接按钮激活状态</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">链接按钮激活状态</a>
```

### 禁用状态

在按钮上设置 `disabled` 属性或者添加 `.am-disabled` class。

`````html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">禁用状态</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">禁用状态</button>
<br><br>
<a href="#" class="am-btn am-btn-primary am-disabled">链接按钮禁用状态</a>
<a href="#" class="am-btn am-btn-default am-disabled">链接按钮禁用状态</a>
`````
```html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">禁用状态</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">禁用状态</button>

<a href="#" class="am-btn am-btn-primary am-disabled">链接按钮禁用状态</a>
<a href="#" class="am-btn am-btn-default am-disabled">链接按钮禁用状态</a>
```

<div class="am-alert am-alert-warning">
  IE9 会把设置了 <code>disabled</code> 属性的 <code>button</code> 元素文字渲染成灰色并加上白色阴影，CSS 无法控制这个默认样式。
</div>


## 按钮尺寸

- `.am-btn-xl`
- `.am-btn-lg`
- `.am-btn-default`
- `.am-btn-sm`
- `.am-btn-xs`

`````html
<button class="am-btn am-btn-default am-btn-xl">按钮 - xl</button>
<button class="am-btn am-btn-default am-btn-lg">按钮 - lg</button>
<button class="am-btn am-btn-default">按钮默认大小</button>
<button class="am-btn am-btn-default am-btn-sm">按钮 - sm</button>
<button class="am-btn am-btn-default am-btn-xs">按钮 - xs</button>
<br />
<br />
<button class="am-btn am-btn-primary am-btn-xl">按钮 - xl</button>
<button class="am-btn am-btn-primary am-btn-lg">按钮 - lg</button>
<button class="am-btn am-btn-primary">按钮默认大小</button>
<button class="am-btn am-btn-primary am-btn-sm">按钮 - sm</button>
<button class="am-btn am-btn-primary am-btn-xs">按钮 - xs</button>
`````
```html
<button class="am-btn am-btn-default am-btn-xl">按钮 - xl</button>
<button class="am-btn am-btn-default am-btn-lg">按钮 - lg</button>
<button class="am-btn am-btn-default">按钮默认大小</button>
<button class="am-btn am-btn-default am-btn-sm">按钮 - sm</button>
<button class="am-btn am-btn-default am-btn-xs">按钮 - xs</button>

<button class="am-btn am-btn-primary am-btn-xl">按钮 - xl</button>
<button class="am-btn am-btn-primary am-btn-lg">按钮 - lg</button>
<button class="am-btn am-btn-primary">按钮默认大小</button>
<button class="am-btn am-btn-primary am-btn-sm">按钮 - sm</button>
<button class="am-btn am-btn-primary am-btn-xs">按钮 - xs</button>
```

## 块级显示

添加 `.am-btn-block` class。

`````html
<button type="button" class="am-btn am-btn-primary am-btn-block">块级显示的按钮</button>
<button type="button" class="am-btn am-btn-default am-btn-block">块级显示的按钮</button>
`````
```html
<button type="button" class="am-btn am-btn-primary am-btn-block">块级显示的按钮</button>
<button type="button" class="am-btn am-btn-default am-btn-block">块级显示的按钮</button>
```

## 按钮 Icon

使用 Icon 之前需先引入 [Icon 组件](/css/icon)。

`````html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  设置
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  结账
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  加载中
</button>

<button class="am-btn am-btn-primary">
  下载片片
  <i class="am-icon-cloud-download"></i>
</button>
`````
```html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  设置
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  结账
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  加载中
</button>

<button class="am-btn am-btn-primary">
  下载片片
  <i class="am-icon-cloud-download"></i>
</button>
```
