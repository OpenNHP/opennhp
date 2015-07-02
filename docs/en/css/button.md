# Button
---

## Usage

### Default Styles

Add the `.am-btn` class to the elements you want to display as a button, and then set its color.

- `.am-btn-default` - Default, gray button
- `.am-btn-primary` - blue button
- `.am-btn-secondary` - light blue button
- `.am-btn-success` - green button
- `.am-btn-warning` - orange button
- `.am-btn-danger` - red button
- `.am-btn-link`

`````html
<button type="button" class="am-btn am-btn-default">Default</button>
<button type="button" class="am-btn am-btn-primary">Primary</button>
<button type="button" class="am-btn am-btn-secondary">Secondary</button>
<button type="button" class="am-btn am-btn-success">Green</button>
<button type="button" class="am-btn am-btn-warning">Orange</button>
<button type="button" class="am-btn am-btn-danger">Red</button>
<a class="am-btn am-btn-link">Link</a>
<a class="am-btn am-btn-danger" href="http://www.bing.com" target="_blank">Link using button style</a>
`````

```html
<button type="button" class="am-btn am-btn-default">Default</button>
<button type="button" class="am-btn am-btn-primary">Primary</button>
<button type="button" class="am-btn am-btn-secondary">Secondary</button>
<button type="button" class="am-btn am-btn-success">Green</button>
<button type="button" class="am-btn am-btn-warning">Orange</button>
<button type="button" class="am-btn am-btn-danger">Red</button>
<a class="am-btn am-btn-link">Link</a>
<a class="am-btn am-btn-danger" href="http://www.bing.com" target="_blank">Link using button style</a>
```

### Round Cornor Button

Add the `.am-radius` class to the default style buttons.

`````html
<button type="button" class="am-btn am-btn-default am-radius">Default</button>
<button type="button" class="am-btn am-btn-primary am-radius">Primary</button>
<button type="button" class="am-btn am-btn-secondary am-radius">Secondary</button>
<button type="button" class="am-btn am-btn-success am-radius">Green</button>
<button type="button" class="am-btn am-btn-warning am-radius">Orange</button>
<button type="button" class="am-btn am-btn-danger am-radius">Red</button>
`````
```html
<button type="button" class="am-btn am-btn-default am-radius">Default</button>
```

### 椭圆形按钮

Add the `.am-round` class to the default style buttons.

`````html
<button type="button" class="am-btn am-btn-default am-round">Default</button>
<button type="button" class="am-btn am-btn-primary am-round">Primary</button>
<button type="button" class="am-btn am-btn-secondary am-round">Secondary</button>
<button type="button" class="am-btn am-btn-success am-round">Green</button>
<button type="button" class="am-btn am-btn-warning am-round">Orange</button>
<button type="button" class="am-btn am-btn-danger am-round">Red</button>
`````
```html
<button type="button" class="am-btn am-btn-default am-round">Default</button>
```

## Button Status

### Active

Add the `.am-active` class to buttons to set button to be active.

`````html
<button type="button" class="am-btn am-btn-primary am-active">Active Button</button>
<button type="button" class="am-btn am-btn-default am-active">Active Button</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">Acitve Link Button</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">Acitve Link Button</a>
`````
```html
<button type="button" class="am-btn am-btn-primary am-active">Active Button</button>
<button type="button" class="am-btn am-btn-default am-active">Active Button</button>
<br >
<br >
<a href="#" class="am-btn am-btn-primary am-active" role="button">Acitve Link Button</a>
<a href="#" class="am-btn am-btn-default am-active" role="button">Acitve Link Button</a>
```

### Disabled

Set the `disabled` attribute or add the `.am-disabled` class to buttons to disable button.

`````html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">Disabled Button</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">Disabled Button</button>
<br><br>
<a href="#" class="am-btn am-btn-primary am-disabled">Disabled Link Button</a>
<a href="#" class="am-btn am-btn-default am-disabled">Disabled Link Button</a>
`````
```html
<button type="button" class="am-btn am-btn-primary" disabled="disabled">Disabled Button</button>
<button type="button" class="am-btn am-btn-default" disabled="disabled">Disabled Button</button>

<a href="#" class="am-btn am-btn-primary am-disabled">Disabled Link Button</a>
<a href="#" class="am-btn am-btn-default am-disabled">Disabled Link Button</a>
```

<div class="am-alert am-alert-warning">
  IE9 will render <code>button</code> with <code>disabled</code> attribute to be gray with white shadow. We can't control this style with CSS.
</div>


## Size

- `.am-btn-xl`
- `.am-btn-lg`
- `.am-btn-default`
- `.am-btn-sm`
- `.am-btn-xs`

`````html
<button class="am-btn am-btn-default am-btn-xl">Button - xl</button>
<button class="am-btn am-btn-default am-btn-lg">Button - lg</button>
<button class="am-btn am-btn-default">Button - default</button>
<button class="am-btn am-btn-default am-btn-sm">Button - sm</button>
<button class="am-btn am-btn-default am-btn-xs">Button - xs</button>
<br />
<br />
<button class="am-btn am-btn-primary am-btn-xl">Button - xl</button>
<button class="am-btn am-btn-primary am-btn-lg">Button - lg</button>
<button class="am-btn am-btn-primary">Button - default</button>
<button class="am-btn am-btn-primary am-btn-sm">Button - sm</button>
<button class="am-btn am-btn-primary am-btn-xs">Button - xs</button>
`````
```html
<button class="am-btn am-btn-default am-btn-xl">Button - xl</button>
<button class="am-btn am-btn-default am-btn-lg">Button - lg</button>
<button class="am-btn am-btn-default">Button - default</button>
<button class="am-btn am-btn-default am-btn-sm">Button - sm</button>
<button class="am-btn am-btn-default am-btn-xs">Button - xs</button>

<button class="am-btn am-btn-primary am-btn-xl">Button - xl</button>
<button class="am-btn am-btn-primary am-btn-lg">Button - lg</button>
<button class="am-btn am-btn-primary">Button - default</button>
<button class="am-btn am-btn-primary am-btn-sm">Button - sm</button>
<button class="am-btn am-btn-primary am-btn-xs">Button - xs</button>
```

## Block

Add the `.am-btn-block` class to display the button in block.

`````html
<button type="button" class="am-btn am-btn-primary am-btn-block">Block Button</button>
<button type="button" class="am-btn am-btn-default am-btn-block">Block Button</button>
`````
```html
<button type="button" class="am-btn am-btn-primary am-btn-block">Block Button</button>
<button type="button" class="am-btn am-btn-default am-btn-block">Block Button</button>
```

## Button Icon

Please import [Icon Plugins](/css/icon) before using Icon.

`````html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  Setting
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  Checkout
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  Loading...
</button>

<button class="am-btn am-btn-primary">
  Download
  <i class="am-icon-cloud-download"></i>
</button>
`````
```html
<button class="am-btn am-btn-default">
  <i class="am-icon-cog"></i>
  Setting
</button>

<a class="am-btn am-btn-warning" href="#">
  <i class="am-icon-shopping-cart"></i>
  Checkout
</a>

<button class="am-btn am-btn-default">
  <i class="am-icon-spinner am-icon-spin"></i>
  Loading...
</button>

<button class="am-btn am-btn-primary">
  Download
  <i class="am-icon-cloud-download"></i>
</button>
```
