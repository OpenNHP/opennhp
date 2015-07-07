---
id: ucheck
title: 选项卡
titleEn: uCheck
prev: javascript/datepicker.html
next: javascript/validator.html
source: js/ui.ucheck.js
doc: docs/javascript/ucheck.md
---

# uCheck
---

A rewrite of `radio` and `checkbox`.

Icon Font is used in uCheck. Please be carefult if your browser don't support Icon Font.

## Example

### Default Style

`````html
<style>
  .am-ucheck-icons [class*="am-icon"] + [class*="am-icon"] {
    margin-left: 0;
  }
</style>

<div class="am-g">
  <div class="am-u-sm-6">
    <h3>Checkbox</h3>
    <label class="am-checkbox needsclick">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox needsclick">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
    <label class="am-checkbox">
      <input type="checkbox" value="" data-am-ucheck disabled>
      Disabled/Not selected.
    </label>
    <label class="am-checkbox">
      <input type="checkbox" checked="checked" value="" data-am-ucheck disabled
             checked>
      Disabled/Selected
    </label>
  </div>

  <div class="am-u-sm-6">
    <h3>Radio</h3>
    <label class="am-radio needsclick">
      <input type="radio" name="radio1" value="" data-am-ucheck>
      Not selected
    </label>
    <label class="am-radio needsclick">
      <input type="radio" name="radio1" value="" data-am-ucheck checked>
      Selected
    </label>

    <label class="am-radio">
      <input type="radio" name="radio2" value="" data-am-ucheck disabled>
      Disabled/Not selected
    </label>
    <label class="am-radio">
      <input type="radio" name="radio2" value="" data-am-ucheck checked
             disabled>
      Disabled/Selected
    </label>
  </div>
</div>
`````
```html
<div class="am-g">
  <div class="am-u-sm-6">
    <h3>Checkbox</h3>
    <label class="am-checkbox needsclick">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox needsclick">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
    <label class="am-checkbox">
      <input type="checkbox" value="" data-am-ucheck disabled>
      Disabled/Not selected.
    </label>
    <label class="am-checkbox">
      <input type="checkbox" checked="checked" value="" data-am-ucheck disabled
             checked>
      Disabled/Selected
    </label>
  </div>

  <div class="am-u-sm-6">
    <h3>Radio</h3>
    <label class="am-radio needsclick">
      <input type="radio" name="radio1" value="" data-am-ucheck>
      Not selected
    </label>
    <label class="am-radio needsclick">
      <input type="radio" name="radio1" value="" data-am-ucheck checked>
      Selected
    </label>

    <label class="am-radio">
      <input type="radio" name="radio2" value="" data-am-ucheck disabled>
      Disabled/Not selected
    </label>
    <label class="am-radio">
      <input type="radio" name="radio2" value="" data-am-ucheck checked
             disabled>
      Disabled/Selected
    </label>
  </div>
</div>
```

### Color

Default color is the primary color. Add color classes to `.am-checkbox`/`.am-checkbox` to change color.

- `.am-secondary`
- `.am-success`
- `.am-warning`
- `.am-danger`

`````html
<div class="am-g">
  <div class="am-u-sm-6 am-u-md-3">
    <h3>Checkbox</h3>
    <label class="am-checkbox am-secondary">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox am-secondary">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Radio</h3>
    <label class="am-radio am-secondary">
      <input type="radio" name="radio3" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-radio am-secondary">
      <input type="radio" name="radio3" value="" data-am-ucheck checked> Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Checkbox</h3>
    <label class="am-checkbox am-success">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox am-success">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Radio</h3>
    <label class="am-radio am-success">
      <input type="radio" name="radio7" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-radio am-success">
      <input type="radio" name="radio7" value="" data-am-ucheck checked> Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Checkbox</h3>
    <label class="am-checkbox am-warning">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox am-warning">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Radio</h3>
    <label class="am-radio am-warning">
      <input type="radio" name="radio4" value="" data-am-ucheck> 未选中
    </label>
    <label class="am-radio am-warning">
      <input type="radio" name="radio4" value="" data-am-ucheck checked> 已选中
    </label>
  </div>

  <!-- red -->
  <div class="am-u-sm-6 am-u-md-3">
    <h3>Checkbox</h3>
    <label class="am-checkbox am-danger">
      <input type="checkbox" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-checkbox am-danger">
      <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
      Selected
    </label>
  </div>

  <div class="am-u-sm-6 am-u-md-3">
    <h3>Radio</h3>
    <label class="am-radio am-danger">
      <input type="radio" name="radio5" value="" data-am-ucheck> Not selected
    </label>
    <label class="am-radio am-danger">
      <input type="radio" name="radio5" value="" data-am-ucheck checked> Selected
    </label>
  </div>
</div>
`````
```html
<label class="am-checkbox am-secondary">
  <input type="checkbox" value="" data-am-ucheck> Not selected
</label>
<label class="am-checkbox am-secondary">
  <input type="checkbox" checked="checked" value="" data-am-ucheck checked>
  Selected
</label>

<label class="am-radio am-secondary">
  <input type="radio" name="radio3" value="" data-am-ucheck> Not selected
</label>
<label class="am-radio am-secondary">
  <input type="radio" name="radio3" value="" data-am-ucheck checked> Selected
</label>
```

<!--### Input Group 中使用

`````html
<div class="am-g">
  <div class="am-u-md-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <label class="am-checkbox">
          <input type="checkbox" data-am-ucheck>
        </label>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-u-md-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <label class="am-radio">
          <input type="radio" data-am-ucheck>
        </label>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
`````
-->

### Inline style

`````html
<div class="am-form-group">
  <h3>Equipments</h3>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> iPhone
  </label>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> iMac
  </label>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> Macbook
  </label>
</div>

<div class="am-form-group">
  <h3>Gender</h3>
  <label class="am-radio-inline">
    <input type="radio" name="radio10" value="male" data-am-ucheck> Male
  </label>
  <label class="am-radio-inline">
    <input type="radio" name="radio10" value="female" data-am-ucheck> Female
  </label>
</div>
`````
```html
<div class="am-form-group">
  <h3>Equipments</h3>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> iPhone
  </label>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> iMac
  </label>
  <label class="am-checkbox-inline">
    <input type="checkbox"  value="" data-am-ucheck> Macbook
  </label>
</div>

<div class="am-form-group">
  <h3>Gender</h3>
  <label class="am-radio-inline">
    <input type="radio" name="radio10" value="male" data-am-ucheck> Male
  </label>
  <label class="am-radio-inline">
    <input type="radio" name="radio10" value="female" data-am-ucheck> Female
  </label>
</div>
```

### Form Validation

`````html
<form class="am-form" data-am-validator>
  <div class="am-form-group">
    <h3>Equipment<sup class="am-text-danger">*</sup> (Select 2 options at least, and 4 at most)</h3>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="ip" data-am-ucheck required minchecked="2" maxchecked="4"> iPhone
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="im" data-am-ucheck> iMac
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="mbp" data-am-ucheck> Macbook Pro
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="sf" data-am-ucheck> Sophie Marceau
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="sur" data-am-ucheck> Surface
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="rb" data-am-ucheck> Razer Blade
    </label>
  </div>

  <div class="am-form-group">
    <h3>Gender <sup class="am-text-danger">*</sup></h3>
    <label class="am-radio">
      <input type="radio" name="radio10" value="male" data-am-ucheck required> Male
    </label>
    <label class="am-radio">
      <input type="radio" name="radio10" value="female" data-am-ucheck> Female
    </label>
    <div>
      <div><button type="submit" class="am-btn am-btn-primary">Submit</button></div>
</form>
`````
```html
<form class="am-form" data-am-validator>
  <div class="am-form-group">
    <h3>Equipment<sup class="am-text-danger">*</sup> (Select 2 options at least, and 4 at most)</h3>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="ip" data-am-ucheck required minchecked="2" maxchecked="4"> iPhone
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="im" data-am-ucheck> iMac
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="mbp" data-am-ucheck> Macbook Pro
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="sf" data-am-ucheck> Sophie Marceau
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="sur" data-am-ucheck> Surface
    </label>
    <label class="am-checkbox">
      <input type="checkbox" name="cbx" value="rb" data-am-ucheck> Razer Blade
    </label>
  </div>

  <div class="am-form-group">
    <h3>Gender <sup class="am-text-danger">*</sup></h3>
    <label class="am-radio">
      <input type="radio" name="radio10" value="male" data-am-ucheck required> Male
    </label>
    <label class="am-radio">
      <input type="radio" name="radio10" value="female" data-am-ucheck> Female
    </label>
    <div>
      <div><button type="submit" class="am-btn am-btn-primary">Submit</button></div>
</form>
```

## Usage

### Using Data API

Add `data-am-ucheck` attribute to `radio`/`checkbox`.

```html
<label class="am-checkbox">
  <input type="checkbox" value="" data-am-ucheck> Not selected
</label>

<label class="am-radio">
  <input type="radio" value="" data-am-ucheck> Not selected
</label>
```

### Using JS

```javascript
$(function() {
  $('input[type='checkbox'], input[type='radio']').uCheck();
});
```

#### Methods

- `$().uCheck('check')`: Check
- `$().uCheck('uncheck')`: Uncheck
- `$().uCheck('toggle')`: Switch between selected and not selected
- `$().uCheck('disable')`: Disable
- `$().uCheck('enable')`: Enable

## Reference

- [iCheck - Highly customizable checkboxes and radio buttons for jQuery and Zepto](https://github.com/fronteed/iCheck)
