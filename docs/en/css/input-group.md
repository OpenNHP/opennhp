# Input Group
-------------

Input group is an extension based on Form and Button.

Add the `.am-input-group` class to container and add the `.am-input-group-label` class to label. See below for detail.

## Usage

### Input and Label

The following code shows how to use icons and texts as labels.

`````html
<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-user am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-lock am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="Password">
</div>

<div class="am-input-group">
  <input type="text" class="am-form-field">
  <span class="am-input-group-label">.00</span>
</div>

<div class="am-input-group">
  <span class="am-input-group-label">$</span>
  <input type="text" class="am-form-field">
  <span class="am-input-group-label">.00</span>
</div>
`````

```html
<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-user am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-lock am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="Password">
</div>

<div class="am-input-group">
  <input type="text" class="am-form-field">
  <span class="am-input-group-label">.00</span>
</div>

<div class="am-input-group">
  <span class="am-input-group-label">$</span>
  <input type="text" class="am-form-field">
  <span class="am-input-group-label">.00</span>
</div>
```

### Radio/Checkbox in Input Group

Use Radio/Checkbox in `.am-input-group-label`.

`````html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="radio">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
`````
```html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="radio">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
```

### Button in Input Group

Use `.am-input-group-btn` instead of `.am-input-group-label`.

`````html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button"><span class="am-icon-search"></span> </button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <input type="text" class="am-form-field">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">手气还行</button>
      </span>
    </div>
  </div>
</div>
`````
```html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button"><span class="am-icon-search"></span> </button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group">
      <input type="text" class="am-form-field">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">手气还行</button>
      </span>
    </div>
  </div>
</div>
```

## Style Modifier

### Size

Add `.am-input-group-lg` or `.am-input-group-sm` to the `.am-input-group`.

`````html
<div class="am-input-group am-input-group-lg">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Large Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Default Username">
</div>

<div class="am-input-group am-input-group-sm">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Small Username">
</div>
`````
```html
<div class="am-input-group am-input-group-lg">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group am-input-group-sm">
  <span class="am-input-group-label">@</span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>
```

### Color

`````html
<div class="am-input-group am-input-group-primary">
  <span class="am-input-group-label"><i class="am-icon-user am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="你的大名">
</div>

<div class="am-input-group am-input-group-secondary">
  <span class="am-input-group-label"><i class="am-icon-credit-card am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="你的银行卡号">
</div>

<div class="am-input-group am-input-group-success">
  <span class="am-input-group-label"><i class="am-icon-money am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="你的银行卡密码">
</div>

<div class="am-input-group am-input-group-warning">
  <span class="am-input-group-label"><i class="am-icon-bank am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="开户行">
</div>

<div class="am-input-group am-input-group-danger">
  <span class="am-input-group-label"><i class="am-icon-location-arrow am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="你所在城市">
</div>
`````
```html
<div class="am-input-group am-input-group-primary">
  <span class="am-input-group-label"><i class="am-icon-user am-icon-fw"></i></span>
  <input type="text" class="am-form-field" placeholder="你的大名">
</div>

<div class="am-input-group am-input-group-secondary">
  ...
</div>

<div class="am-input-group am-input-group-success">
  ...
</div>

<div class="am-input-group am-input-group-warning">
  ...
</div>

<div class="am-input-group am-input-group-danger">
  ...
</div>
```

When changing color for buttons, we need to change the style of both button and container.

`````html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group am-input-group-danger">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group am-input-group-primary">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-primary" type="button"><span class="am-icon-search"></span></button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-lg-6">
    <div class="am-input-group am-input-group-danger">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-lg-6">
    <div class="am-input-group am-input-group-primary">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-primary" type="button"><span class="am-icon-search"></span></button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
```
