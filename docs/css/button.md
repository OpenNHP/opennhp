# Button
---

## 基本使用

在要应用按钮样式的元素上添加 `.am-btn`，再设置相应的颜色。

- `.am-btn-default` - 默认，灰色按钮
- `.am-btn-primary` - 蓝色按钮
- `.am-btn-secondary` - 浅蓝色按钮
- `.am-btn-success` - 绿色按钮
- `.am-btn-warning` - 橙色按钮
- `.am-btn-danger` - 红色按钮
- `.am-btn-link`

`````html
<button type="button" class="am-btn am-btn-default">Default</button>
<button type="button" class="am-btn am-btn-primary">Primary</button>
<button type="button" class="am-btn am-btn-secondary">Secondary</button>
<button type="button" class="am-btn am-btn-success">Success</button>
<button type="button" class="am-btn am-btn-warning">Warning</button>
<button type="button" class="am-btn am-btn-danger">Danger</button>
<button type="button" class="am-btn am-btn-link">Link</button>
<button type="button" class="am-btn am-btn-default">Link button like style</button>
`````

```html
<button type="button" class="am-btn am-btn-default">Default</button>
<button type="button" class="am-btn am-btn-primary">Primary</button>
<button type="button" class="am-btn am-btn-secondary">Secondary</button>
<button type="button" class="am-btn am-btn-success">Success</button>
<button type="button" class="am-btn am-btn-warning">Warning</button>
<button type="button" class="am-btn am-btn-danger">Danger</button>
<button type="button" class="am-btn am-btn-link">Link</button>
<button type="button" class="am-btn am-btn-default">Link button like style</button>
```

## 按钮状态

### 激活状态

在按钮上添加 `.am-active` class。

`````html
<button type="button" class="am-btn am-btn-primary am-active">Primary button</button>
<button type="button" class="am-btn am-btn-default am-active">Button</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">Primary link</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">Link</a>
`````
```html
<button type="button" class="am-btn am-btn-primary am-active">Primary button</button>
<button type="button" class="am-btn am-btn-default am-active">Button</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">Primary link</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">Link</a>
```

### 禁用状态

在按钮上设置 `disabled` 属性或者添加 `.am-disabled` class。

`````html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">Primary button</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">Button</button>
<br><br>
<a href="#" class="am-btn am-btn-primary am-disabled">Primary link</a>
<a href="#" class="am-btn am-btn-default am-disabled">Link</a>
`````
```html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">Primary button</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">Button</button>

<a href="#" class="am-btn am-btn-primary am-disabled">Primary link</a>
<a href="#" class="am-btn am-btn-default am-disabled">Link</a>
```

<div class="am-alert am-alert-warning">
  IE9 会把设置了 <code>disabled</code> 属性的 <code>button</code> 元素文字渲染成灰色并加上白色阴影，CSS 无法控制这个默认样式。
</div>


## 按钮尺寸

- `.am-btn-lg`
- `.am-btn-default`
- `.am-btn-sm`
- `.am-btn-xs`

`````html
<button type="button" class="am-btn am-btn-default am-btn-lg">Large button</button>
<button type="button" class="am-btn am-btn-default">Default button</button>
<button type="button" class="am-btn am-btn-default am-btn-sm">Small button</button>
<button type="button" class="am-btn am-btn-default am-btn-xs">Extra small button</button>
<br />
<br />
<button type="button" class="am-btn am-btn-primary am-btn-lg">Large button</button>
<button type="button" class="am-btn am-btn-primary">Default button</button>
<button type="button" class="am-btn am-btn-primary am-btn-sm">Small button</button>
<button type="button" class="am-btn am-btn-primary am-btn-xs">Extra small button</button>
`````
```html
<button type="button" class="am-btn am-btn-default am-btn-lg">Large button</button>
<button type="button" class="am-btn am-btn-default">Default button</button>
<button type="button" class="am-btn am-btn-default am-btn-sm">Small button</button>
<button type="button" class="am-btn am-btn-default am-btn-xs">Extra small button</button>

<button type="button" class="am-btn am-btn-primary am-btn-lg">Large button</button>
<button type="button" class="am-btn am-btn-primary">Default button</button>
<button type="button" class="am-btn am-btn-primary am-btn-sm">Small button</button>
<button type="button" class="am-btn am-btn-primary am-btn-xs">Extra small button</button>
```


## 块级显示

添加 `.am-btn-block` class。

`````html
<button type="button" class="am-btn am-btn-primary am-btn-block">Block level button</button>
<button type="button" class="am-btn am-btn-default am-btn-block">Block level button</button>
`````
```html
<button type="button" class="am-btn am-btn-primary am-btn-block">Block level button</button>
<button type="button" class="am-btn am-btn-default am-btn-block">Block level button</button>
```

## 按钮 Icon

使用 Icon 之前需先引入 [Icon 组件](/css/icon)。

`````html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  Settings
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  Checkout
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  Loading
</button>

<button class="am-btn am-btn-primary">
  Download
  <i class="am-icon-cloud-download"></i>
</button>
`````
```html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  Settings
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  Checkout
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  Loading
</button>

<button class="am-btn am-btn-primary">
  Download
  <i class="am-icon-cloud-download"></i>
</button>
```