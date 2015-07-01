# Button-group
---

A Group of buttons.

## Usage

Put a series of `.am-btn` buttons into `.am-btn-group`.

`````html
<div class="am-btn-group">
    <button type="button" class="am-btn am-btn-default">Left Hand</button>
    <button type="button" class="am-btn am-btn-default">Doge's Hand</button>
    <button type="button" class="am-btn am-btn-default">Right Hand</button>
</div>

&nbsp;

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-radius">Left Hand</button>
  <button type="button" class="am-btn am-btn-primary am-radius">Doge's Hand</button>
  <button type="button" class="am-btn am-btn-primary am-radius">Right Hand</button>
</div>

&nbsp;

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-success am-round">Left Hand</button>
  <button type="button" class="am-btn am-btn-success am-round">Doge's Hand</button>
  <button type="button" class="am-btn am-btn-success am-round">Right Hand</button>
</div>
`````

```html
<div class="am-btn-group">
 <button type="button" class="am-btn am-btn-default">Left Hand</button>
  ...
</div>

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-radius">Left Hand</button>
  ...
</div>

<div class="am-btn-group">
  <button type="button" class="am-btn am-btn-primary am-round">Left Hand</button>
  ...
</div>
`````

## Button Tool Bar

Put `.am-btn-group` into `.am-btn-toolbar` to make a button tool bar.


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

## Size of buttons

Add `.am-btn-group-lg`, `.am-btn-group-sm` or `.am-btn-group-xs` to `.am-btn-group ` to change the size of bottons.


`````html
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-lg">
        <button type="button" class="am-btn am-btn-default">Left Hand lg</button>
        <button type="button" class="am-btn am-btn-default">Doge's Hand lg</button>
        <button type="button" class="am-btn am-btn-default">Right Hand lg</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group">
        <button type="button" class="am-btn am-btn-default">Left Hand Default</button>
        <button type="button" class="am-btn am-btn-default">Doge's Hand Default</button>
        <button type="button" class="am-btn am-btn-default">Right Hand Default</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-sm">
        <button type="button" class="am-btn am-btn-default">Left Hand sm</button>
        <button type="button" class="am-btn am-btn-default">Doge's Hand sm</button>
        <button type="button" class="am-btn am-btn-default">Left Hand sm</button>
    </div>
</div>
<br/>
<div class="am-btn-toolbar">
    <div class="am-btn-group am-btn-group-xs">
        <button type="button" class="am-btn am-btn-default">Left Hand xs</button>
        <button type="button" class="am-btn am-btn-default">Doge's Hand xs</button>
        <button type="button" class="am-btn am-btn-default">Left Hand xs</button>
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

## Vertical Arrangement

Use `.am-btn-group-stacked` to arrange buttons vertically.

`````html
<div class="am-btn-group-stacked">
    <button type="button" class="am-btn am-btn-default">This is a button</button>
    <button type="button" class="am-btn am-btn-default">This is a button</button>
    <button type="button" class="am-btn am-btn-default">This is a button</button>
    <button type="button" class="am-btn am-btn-default">This is a button</button>
</div>
`````
```html
<div class="am-btn-group-stacked">
  <button type="button" class="am-btn am-btn-default">Button</button>
  ...
</div>
```

## Justify button group

Add `.am-btn-group-justify` to justify a button group.

~~**Attention: ** This can only be apply to `<a>` element and `<button>` element. Don't works for `display: table-cell` style~~.

**Implemented using `flexbox`, so this only works on IE 10+ and other modern browser**.

`````html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">Left Hand</a>
  <a class="am-btn am-btn-default" role="button">Doge's Hand</a>
  <a class="am-btn am-btn-default" role="button">Right Hand</a>
</div>
`````
```html
<div class="am-btn-group am-btn-group-justify">
  <a class="am-btn am-btn-default" role="button">Left Hand</a>
  ...
</div>
```

## Combine with Dropdown plugin

The following example shows how to combine button-group with [Dropdown](/javascript/dropdown?_ver=2.x).

### Example(Down)

`````html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">Down</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
        <li class="am-dropdown-header">Title</li>
        <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
        <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
        <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
        <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
        <li class="am-divider"></li>
        <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
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
  <button class="am-btn am-btn-secondary">Down</button>
  <div class="am-dropdown" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-down"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Title</li>
      <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
      <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
      <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
      <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
      <li class="am-divider"></li>
      <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
    </ul>
  </div>
</div>
```

### Example (Up)

`````html
<div class="am-btn-group">
  <button class="am-btn am-btn-secondary">Up</button>
  <div class="am-dropdown am-dropdown-up" data-am-dropdown>
    <button class="am-btn am-btn-secondary am-dropdown-toggle" data-am-dropdown-toggle> <span class="am-icon-caret-up"></span></button>
    <ul class="am-dropdown-content">
      <li class="am-dropdown-header">Title</li>
      <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
      <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
      <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
      <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
      <li class="am-divider"></li>
      <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
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
      <li class="am-dropdown-header">Title</li>
      <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
      <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
      <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
      <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
      <li class="am-divider"></li>
      <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
    </ul>
  </div>
</div>
```
