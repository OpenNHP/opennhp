# Form
---

Easily create nicely looking forms with different styles and layouts.

## Usage


### Radio and Checkbox

`<input>` with `checkbox`、`radio` type is slightly different from other elements：

- Add `.am-checkbox`、`.am-radio` class to container when display in block
- Add `.am-checkbox-inline`、`.am-radio-inline` class to container when display inline.

### Select

We can hardly define cross-browser styles for `select` simply with CSS, so keep its default style can probably the best solution.([This is what they did in Pure CSS](http://purecss.io/forms/#stacked-form)). Amaze UI provide some styles for Webkit browsers to replace the default arrows.

### File Selector

`<input type="file">` is another problem CSS can't handle. If you don't like the default style, you can just set its opacity to 0. and cover other elements with it.

`````html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-default am-btn-sm">
    <i class="am-icon-cloud-upload"></i> Upload File</button>
  <input type="file" multiple>
</div>

<hr/>

<div class="am-form-group am-form-file">
  <i class="am-icon-cloud-upload"></i> Upload File
  <input type="file" multiple>
</div>
`````

```html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-default am-btn-sm">
    <i class="am-icon-cloud-upload"></i> Upload File</button>
  <input type="file" multiple>
</div>

<hr/>

<div class="am-form-group am-form-file">
  <i class="am-icon-cloud-upload"></i> Upload File
  <input type="file" multiple>
</div>
```

The problem is the selected files won't be shown. Please [use it JS](https://developer.mozilla.org/en-US/docs/Web/API/FileList).

`````html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-danger am-btn-sm">
    <i class="am-icon-cloud-upload"></i> Upload File</button>
  <input id="doc-form-file" type="file" multiple>
</div>
<div id="file-list"></div>
<script>
  $(function() {
    $('#doc-form-file').on('change', function() {
      var fileNames = '';
      $.each(this.files, function() {
        fileNames += '<span class="am-badge">' + this.name + '</span> ';
      });
      $('#file-list').html(fileNames);
    });
  });
</script>
`````
```html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-danger am-btn-sm">
    <i class="am-icon-cloud-upload"></i> Upload File</button>
  <input id="doc-form-file" type="file" multiple>
</div>
<div id="file-list"></div>
<script>
  $(function() {
    $('#doc-form-file').on('change', function() {
      var fileNames = '';
      $.each(this.files, function() {
        fileNames += '<span class="am-badge">' + this.name + '</span> ';
      });
      $('#file-list').html(fileNames);
    });
  });
</script>
```

### Example

The styles defined in Amaze UI will be applied to elements only if `.am-form` class is added to container.

`````html
<form class="am-form">
  <fieldset>
    <legend>Title</legend>

    <div class="am-form-group">
      <label for="doc-ipt-email-1">Email</label>
      <input type="email" class="" id="doc-ipt-email-1" placeholder="input your email here">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-pwd-1">password</label>
      <input type="password" class="" id="doc-ipt-pwd-1" placeholder="input your password here">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-file-1">Default file upload zone</label>
      <input type="file" id="doc-ipt-file-1">
      <p class="am-form-help">Please select files to upload...</p>
    </div>

    <div class="am-form-group am-form-file">
      <label for="doc-ipt-file-2">Amaze UI file upload zone</label>
      <div>
        <button type="button" class="am-btn am-btn-default am-btn-sm">
          <i class="am-icon-cloud-upload"></i> Upload File</button>
      </div>
      <input type="file" id="doc-ipt-file-2">
    </div>

    <hr/>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> Checkbox
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option1" checked>
        Radio - option 1
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option2">
        Radio - option 2
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option1"> select me
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option2"> or select me
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option3"> or me
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docInlineRadio"> every minute
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> every second
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> is nice
      </label>
    </div>

    <div class="am-form-group am-form-select">
      <label for="doc-select-1">select</label>
      <select id="doc-select-1">
        <option value="option1">option 1...</option>
        <option value="option2">option 2.....</option>
        <option value="option3">option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">checkbox</label>
      <select multiple class="" id="doc-select-2">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-ta-1">text</label>
      <textarea class="" rows="5" id="doc-ta-1"></textarea>
    </div>

    <p><button type="submit" class="am-btn am-btn-default">submit</button></p>
  </fieldset>
</form>
`````

```html
<form class="am-form">
  <fieldset>
    <legend>Title</legend>

    <div class="am-form-group">
      <label for="doc-ipt-email-1">Email</label>
      <input type="email" class="" id="doc-ipt-email-1" placeholder="input your email here">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-pwd-1">password</label>
      <input type="password" class="" id="doc-ipt-pwd-1" placeholder="input your password here">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-file-1">Default file upload zone</label>
      <input type="file" id="doc-ipt-file-1">
      <p class="am-form-help">Please select files to upload...</p>
    </div>

    <div class="am-form-group am-form-file">
      <label for="doc-ipt-file-2">Amaze UI file upload zone</label>
      <div>
        <button type="button" class="am-btn am-btn-default am-btn-sm">
          <i class="am-icon-cloud-upload"></i> Upload File</button>
      </div>
      <input type="file" id="doc-ipt-file-2">
    </div>

    <hr/>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> Checkbox
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option1" checked>
        Radio - option 1
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option2">
        Radio - option 2
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option1"> select me
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option2"> or select me
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option3"> or me
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docInlineRadio"> every minute
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> every second
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> is nice
      </label>
    </div>

    <div class="am-form-group am-form-select">
      <label for="doc-select-1">select</label>
      <select id="doc-select-1">
        <option value="option1">option 1...</option>
        <option value="option2">option 2.....</option>
        <option value="option3">option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">checkbox</label>
      <select multiple class="" id="doc-select-2">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-ta-1">textbox</label>
      <textarea class="" rows="5" id="doc-ta-1"></textarea>
    </div>

    <p><button type="submit" class="am-btn am-btn-default">submit</button></p>
  </fieldset>
</form>
```

### Form Shape

`````html
<p><input type="text" class="am-form-field am-radius" placeholder="round cornor form" /></p>
<p><input type="text" class="am-form-field am-round" placeholder="ellipse form"/></p>
`````
```html
<p><input type="text" class="am-form-field am-radius" placeholder="round cornor form" /></p>
<p><input type="text" class="am-form-field am-round" placeholder="ellipse form"/></p>
```

## Form Status

Set different status for form.

### Disable Single Element

Add the `disabled` to `<input>` to disable form elements.

`````html
<form class="am-form">
  <input class="am-form-field" type="text" placeholder="Disabled form..." disabled>
</form>
`````

```html
<form class="am-form">
  <input class="am-form-field" type="text" placeholder="Disabled form..." disabled>
</form>
```

### Disable Elements in Field

Add the `disabled` to `<fieldset>` to disable all child elements.

`````html
<form class="am-form">
  <fieldset disabled>
    <div class="am-form-group">
      <label for="doc-ds-ipt-1">Disabled textbox</label>
      <input type="text" id="doc-ds-ipt-1" class="am-form-field" placeholder="Disabled">
    </div>

    <div class="am-form-group">
      <label for="oc-ds-select-1">Disabled Select</label>
      <select id="oc-ds-select-1" class="am-form-field">
        <option>Disabled</option>
      </select>
    </div>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> Disabled checkbox
      </label>
    </div>

    <button type="submit" class="am-btn am-btn-primary">Submit</button>
  </fieldset>
</form>
`````

```html
<form class="am-form">
  <fieldset disabled>
    <div class="am-form-group">
      <label for="doc-ds-ipt-1">Disabled textbox</label>
      <input type="text" id="doc-ds-ipt-1" class="am-form-field" placeholder="Disabled">
    </div>

    <div class="am-form-group">
      <label for="oc-ds-select-1">Disabled Select</label>
      <select id="oc-ds-select-1" class="am-form-field">
        <option>Disabled</option>
      </select>
    </div>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> Disabled checkbox
      </label>
    </div>

    <button type="submit" class="am-btn am-btn-primary">Submit</button>
  </fieldset>
</form>
```

### Disable Link

Add the `.am-disabled` class to `<a>` element to disable link.

`````html
<a class="am-btn am-btn-primary am-disabled">Disabled</a>
`````
```html
<a class="am-btn am-btn-primary am-disabled">Disabled</a>
```

## Layout

### Vertical Layout

Add the `.am-form-horizontal` class to `<form>` and use with Grid.

`````html
<form class="am-form am-form-horizontal">
  <div class="am-form-group">
    <label for="doc-ipt-3" class="am-u-sm-2 am-form-label">Email</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3" placeholder="Please input your email">
    </div>
  </div>

  <div class="am-form-group">
    <label for="doc-ipt-pwd-2" class="am-u-sm-2 am-form-label">Password</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-2" placeholder="Please input your password">
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="checkbox">
        <label>
          <input type="checkbox"> remember me
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">Submit</button>
    </div>
  </div>
</form>
`````

```html
<form class="am-form am-form-horizontal">
  <div class="am-form-group">
    <label for="doc-ipt-3" class="am-u-sm-2 am-form-label">Email</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3" placeholder="Please input your email">
    </div>
  </div>

  <div class="am-form-group">
    <label for="doc-ipt-pwd-2" class="am-u-sm-2 am-form-label">Password</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-2" placeholder="Please input your password">
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="checkbox">
        <label>
          <input type="checkbox"> remember me
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">Submit</button>
    </div>
  </div>
</form>
```

### Inline Layout

Add the `.am-form-inline` class to the container. **Attention**: Right margin is not set for inline elements. Default margin is same as `inline-block`. Compressing HTML will cause the right margin to disappear.

`````html
<form class="am-form-inline" role="form">
  <div class="am-form-group">
    <input type="email" class="am-form-field" placeholder="Email">
  </div>

  &nbsp;

  <div class="am-form-group">
    <input type="password" class="am-form-field" placeholder="Password">
  </div>

  &nbsp;

  <div class="am-checkbox">
    <label>
      <input type="checkbox"> Remember me
    </label>
  </div>

  &nbsp;

  <button type="submit" class="am-btn am-btn-default">Login</button>
</form>
`````
```html
<form class="am-form-inline" role="form">
  <div class="am-form-group">
    <input type="email" class="am-form-field" placeholder="Email">
  </div>
  <div class="am-form-group">
    <input type="password" class="am-form-field" placeholder="Password">
  </div>
  <div class="am-checkbox">
    <label>
      <input type="checkbox"> Remember me
    </label>
  </div>
  <button type="submit" class="am-btn am-btn-default">Login</button>
</form>
```

## Form Icon

Add the `.am-form-icon` to group elements. Need `icon` component.

`````html
<form action="" class="am-form am-form-inline">
  <div class="am-form-group am-form-icon">
    <i class="am-icon-calendar"></i>
    <input type="text" class="am-form-field" placeholder="Date">
  </div>
  &nbsp;
  <div class="am-form-group am-form-icon">
    <i class="am-icon-clock-o"></i>
    <input type="text" class="am-form-field" placeholder="Time">
  </div>
</form>
`````

```html
<form action="" class="am-form am-form-inline">
  <div class="am-form-group am-form-icon">
    <i class="am-icon-calendar"></i>
    <input type="text" class="am-form-field" placeholder="Date">
  </div>

  <div class="am-form-group am-form-icon">
    <i class="am-icon-clock-o"></i>
    <input type="text" class="am-form-field" placeholder="Time">
  </div>
</form>
```

## Verification Status

### Example

`````html
<form action="" class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label class="am-form-label" for="doc-ipt-success">Successful Verification</label>
    <input type="text" id="doc-ipt-success" class="am-form-field">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning">
    <label class="am-form-label" for="doc-ipt-warning">Warning Verification</label>
    <input type="text" id="doc-ipt-warning" class="am-form-field">
  </div>
  <div class="am-form-group am-form-error">
    <label class="am-form-label" for="doc-ipt-error">Error Verification</label>
    <input type="text" id="doc-ipt-error" class="am-form-field">
  </div>
</form>
`````

```html
<form action="" class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label class="am-form-label" for="doc-ipt-success">Successful Verification</label>
    <input type="text" id="doc-ipt-success" class="am-form-field">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning">
    <label class="am-form-label" for="doc-ipt-warning">Warning Verification</label>
    <input type="text" id="doc-ipt-warning" class="am-form-field">
  </div>
  <div class="am-form-group am-form-error">
    <label class="am-form-label" for="doc-ipt-error">Error Verification</label>
    <input type="text" id="doc-ipt-error" class="am-form-field">
  </div>
</form>
```
### Verification with Icons

Add `.am-form-icon` and `.am-form-feedback`。

~~Attention: Styles of icon is written for single line `.am-form-group`. Problems may occur when using muli-line `.am-form-group`.~~

Style is adjusted in `v2.3.1`. Now form with `label` is also supported.

`````html
<form class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Successful Verification">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Warning Verification">
    <span class="am-icon-warning"></span>
  </div>
  <div class="am-form-group am-form-error am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Error Verification">
    <span class="am-icon-times"></span>
  </div>
</form>
`````

```html
<form class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Successful Verification">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Warning Verification">
    <span class="am-icon-warning"></span>
  </div>
  <div class="am-form-group am-form-error am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="Error Verification">
    <span class="am-icon-times"></span>
  </div>
</form>
```

**水平排列：**

`````html
<form class="am-form am-form-horizontal">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label for="doc-ipt-3-a" class="am-u-sm-2 am-form-label">电子邮件</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-a" class="am-form-field" placeholder="输入你的电子邮件">
      <span class="am-icon-warning"></span>
    </div>
  </div>
</form>
`````

```html
<form class="am-form am-form-horizontal">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label for="doc-ipt-3-a" class="am-u-sm-2 am-form-label">电子邮件</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-a" class="am-form-field" placeholder="输入你的电子邮件">
      <span class="am-icon-warning"></span>
    </div>
  </div>
</form>
```



## Size of Form

### Size of Single Element

Add following classes to elements:

- `am-input-lg`
- `am-input-sm`

**Attention: **This is only for forms without `<label>`.

`````html
<form class="am-form">
  <input class="am-form-field am-input-lg" type="text" placeholder="Input with .am-input-lg">
  <br/>
  <input class="am-form-field" type="text" placeholder="Default input">
  <br/>
  <input class="am-form-field am-input-sm" type="text" placeholder="Input with .am-input-sm">
  <br/>

  <div class="am-form-group am-form-select">
    <select class=" am-input-lg">
      <option value="">Select with .am-input-lg</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class="">
      <option value="">Default select</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class=" am-input-sm">
      <option value="">Select with .am-input-sm</option>
    </select>
  </div>
</form>
`````

```html
<form class="am-form">
  <input class="am-form-field am-input-lg" type="text" placeholder="Input with .am-input-lg">
  <br/>
  <input class="am-form-field" type="text" placeholder="Default input">
  <br/>
  <input class="am-form-field am-input-sm" type="text" placeholder="Input with .am-input-sm">
  <br/>

  <div class="am-form-group am-form-select">
    <select class=" am-input-lg">
      <option value="">Select with .am-input-lg</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class="">
      <option value="">Default select</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class=" am-input-sm">
      <option value="">Select with .am-input-sm</option>
    </select>
  </div>
</form>
```

### Size of Group

Size of form can also be adjusted by adding following classes to `.am-form-group`: 

- `.am-form-group-sm`
- `.am-form-group-lg`

Attention: **The `.am-form-field` class should be added to `input` element instead of classes like `.am-input-sm`. 

`````html
<form class="am-form am-form-horizontal">
  <div class="am-form-group am-form-group-sm">
    <label for="doc-ipt-3-1" class="am-u-sm-2 am-form-label">Email</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-1" class="am-form-field" placeholder="Input email here">
    </div>
  </div>

  <div class="am-form-group am-form-group-lg">
    <label for="doc-ipt-pwd-21" class="am-u-sm-2 am-form-label">Password</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-21" class="am-form-field" placeholder="Input password here">
    </div>
  </div>

  <div class="am-form-group am-form-group-sm">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="am-checkbox">
        <label>
          <input type="checkbox"> Remember me
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">Submit</button>
    </div>
  </div>
</form>
`````

```html
<form class="am-form am-form-horizontal">
  <!-- am-form-group 的基础上添加了 am-form-group-sm -->
  <div class="am-form-group am-form-group-sm">
    <label for="doc-ipt-3-1" class="am-u-sm-2 am-form-label">Email</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-1" class="am-form-field" placeholder="Input email here">
    </div>
  </div>

  <div class="am-form-group am-form-group-lg">
    <label for="doc-ipt-pwd-21" class="am-u-sm-2 am-form-label">Password</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-21" class="am-form-field" placeholder="Input password here">
    </div>
  </div>

  <div class="am-form-group am-form-group-sm">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="am-checkbox">
        <label>
          <input type="checkbox"> Remember me
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">Submit</button>
    </div>
  </div>
</form>
```

## Input Set

Use `.am-form-set` to group `<input>` elements to be a set.

`````html
<div class="am-g">
  <div class="am-u-md-8 am-u-sm-centered">
    <form class="am-form">
      <fieldset class="am-form-set">
        <input type="text" placeholder="Your Name">
        <input type="text" placeholder="Password">
        <input type="email" placeholder="Email">
      </fieldset>
      <button type="submit" class="am-btn am-btn-primary am-btn-block">Register</button>
    </form>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-md-8 am-u-sm-centered">
    <form class="am-form">
      <fieldset class="am-form-set">
        <input type="text" placeholder="Your Name">
        <input type="text" placeholder="Password">
        <input type="email" placeholder="Email">
      </fieldset>
      <button type="submit" class="am-btn am-btn-primary am-btn-block">Register</button>
    </form>
  </div>
</div>
```
## Reference

- [Pure CSS styles for Radio/Checkbox in Webkit browser](http://jsbin.com/gitovovidu)
