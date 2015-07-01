# Nav
---

Navigator Component. Add the `.am-nav` class to `<ul>` list.

## Default Style

Add the `.am-nav` class to `<ul>` list.
This is a basic vertical navigator, which can be used with grid.

`````html
<ul class="am-nav">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
`````
```html
<ul class="am-nav">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
```

## Horizontal Navigator

Add the `.am-nav-pills` class to `.am-nav` to make a horizontal navigator.

`````html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
`````
```html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
```

## Navigation Tab

Add the `.am-nav-tabs` class to `.am-nav` to form a navigation tab. Add `.am-active` for activated tab.

`````html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
`````
```html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Getting Started</a></li>
  <li><a href="#">Customize</a></li>
</ul>
```

## Width Justify

Add the `.am-nav-justify` class to make `<li>`s have the same width.

This only work for `media-up` (> 640px). The tabs will arranged vertically when viewport is smaller than 640px.

`````html
<ul class="am-nav am-nav-pills am-nav-justify">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">Getting Started</a></li>
    <li><a href="#">Customize</a></li>
    <li><a href="#">Join us</a></li>
</ul>

<ul class="am-nav am-nav-tabs am-nav-justify">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">Getting Started</a></li>
    <li><a href="#">Join us</a></li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills am-nav-justify">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">Getting Started</a></li>
    <li><a href="#">Customize</a></li>
    <li><a href="#">Join us</a></li>
</ul>

<ul class="am-nav am-nav-tabs am-nav-justify">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">Getting Started</a></li>
    <li><a href="#">Join us</a></li>
</ul>
```

## Status

These status classes can be applied to `<li>`.

- `.am-disabled` - disabled
- `.am-active` - activated

`````html
<ul class="am-nav am-nav-pills">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">About us</a></li>
    <li class="am-disabled"><a href="#">Disabled</a></li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills">
    <li class="am-active"><a href="#">Home</a></li>
    <li><a href="#">About us</a></li>
    <li class="am-disabled"><a href="#">Disabled</a></li>
</ul>
```


## Title and Divider

Title and divider can only be used in vertical navigator.

- `.am-nav-header` Title. Add to `<li>`.
- `.am-nav-divider` Divider, Add to empty `<li>`.

`````html
<ul class="am-nav">
  <li><a href="#"><span class="am-icon-home"></span>Home</a></li>
  <li class="am-nav-header">Getting Started</li>
  <li><a href="#">About us</a></li>
  <li><a href="#">Contact us</a></li>
  <li class="am-nav-divider"></li>
  <li><a href="#">Responsive</a></li>
  <li><a href="#">Mobile first</a></li>
</ul>
`````

```html
<ul class="am-nav">
  <li><a href="#"><span class="am-icon-home"></span>Home</a></li>
  <li class="am-nav-header">Getting Started</li>
  <li><a href="#">About us</a></li>
  <li><a href="#">Contact us</a></li>
  <li class="am-nav-divider"></li>
  <li><a href="#">Responsive</a></li>
  <li><a href="#">Mobile first</a></li>
</ul>
```


## Dropdown

Combine with JS [Dropdown](/##) component.

`````html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Project</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      Menu <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. One line of code</a></li>
      <li class="am-active"><a href="#">2. Don't change URL</a></li>
      <li><a href="#">3. Runtime Sync</a></li>
      <li class="am-disabled"><a href="#">4. Cross-platform adaption on cloud</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. Your exclusive quick dial</a></li>
    </ul>
  </li>
</ul>
`````

```html
<ul class="am-nav am-nav-pills">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Project</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      Menu <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. One line of code</a></li>
      <li class="am-active"><a href="#">2. Don't change URL</a></li>
      <li><a href="#">3. Runtime Sync</a></li>
      <li class="am-disabled"><a href="#">4. Cross-platform adaption on cloud</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. Your exclusive quick dial</a></li>
    </ul>
  </li>
</ul>
```


### Tab


`````html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Project</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      Menu <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Header</li>
      <li><a href="#">1. One line of code</a></li>
      <li class="am-active"><a href="#">2. Don't change URL</a></li>
      <li><a href="#">3. Runtime Sync</a></li>
      <li class="am-disabled"><a href="#">4. Cross-platform adaption on cloud</a></li>
      <li class="am-divider"></li>
      <li><a href="#">5. Your exclusive quick dial</a></li>
    </ul>
  </li>
</ul>
`````

```html
<ul class="am-nav am-nav-tabs">
  <li class="am-active"><a href="#">Home</a></li>
  <li><a href="#">Project</a></li>
  <li class="am-dropdown" data-am-dropdown>
    <a class="am-dropdown-toggle" data-am-dropdown-toggle href="javascript:;">
      Menu <span class="am-icon-caret-down"></span>
    </a>
    <ul class="am-dropdown-content">
      ...
    </ul>
  </li>
</ul>
```
