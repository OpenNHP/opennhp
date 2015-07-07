---
id: validator
title: 表单验证
titleEn: Validator
prev: javascript/ucheck.html
next: javascript/cookie.html
source: js/ui.validator.js
doc: docs/javascript/validator.md
---

# Form Validator
---

Based on HTML 5 form validation, form validator uses attributes in HTML 5 to validate the form, such as `type`, `required`, `pattern`, `min`, `max`, `minlength` and `maxlength`. Form validator can still work like normal H5 validation in the enviroment that don't support JS.

## Examples

### HTML5 Native Form Validation

If the form is only for HTML5 and don't need control, there is no doubt that the native validation is the best solution, and styles can be controlled using pseudo classes such as `:valid` and `:invalid`.

`````html
<form action="" class="am-form">
  <fieldset>
    <legend>H5 Native Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-1">Username：</label>
      <input type="text" id="doc-vld-name-1" maxlength="3" pattern="^\d+$" placeholder="Input username here" class="am-form-field" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-1">Email：</label>
      <input type="email" id="doc-vld-email-1" placeholder="Input email here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-1">URL：</label>
      <input type="url" id="doc-vld-url-1" placeholder="Input URL here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-1">Age：</label>
      <input type="number" class=""  id="doc-vld-age-1" max="100" placeholder="Input age here" required />
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-1">Comment：</label>
      <textarea id="doc-vld-ta-1" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form">
  <fieldset>
    <legend>H5 Native Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-1">Username：</label>
      <input type="text" id="doc-vld-name-1" maxlength="3" pattern="^\d+$" placeholder="Input username here" class="am-form-field" required/>
    </div>
  </fieldset>
</form>
```

**Reference:**

- [Forms in HTML](https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/Forms_in_HTML)
- [:invalid pseudo class](https://developer.mozilla.org/en-US/docs/Web/CSS/:invalid)
- [HTML5 Form Validation](http://www.sitepoint.com/html5-form-validation/)
- [HTML5 Form Validation Examples](http://www.the-art-of-web.com/html/html5-form-validation/)

### JS Form Validation

JS form validation is based on HTML5 form validation:

- `required`: Can't be empty;
- `pattern`: Test with regular expression. Regular expression for `email`, `url` and `number` are provided by plugin;
- `minlength`/`maxlength`: Maximum length;
- `min`/`max`: Maximum, Minimum of value. Only avaliable in number-type input boxes;
- `minchecked`/`maxchecked`: Maximum, minimum number of checked boxes. Avaliable in `checkbox` and dropdown. The relative attributes are set in the first element when using `checkbox`;
- `.js-pattern-xx`: Validation rule classes. Rules in regular expression libraries can simply be validated by adding corresponding classes.

**Attention:**

`pattern` in HTML5 only validate if the input is legal, which means it can be legal expression or empty. If this value is required, please add `required` attribute.

```html
<!-- The following validations are equivalent -->
<input type="email"/>

<!-- xx in js-pattern-xx is the key in the pattern library -->
<input type="text" class="js-pattern-email"/>

<input type="text" pattern="^(...email regex...)$"/>
```

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>JS Form Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username: </label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">Email: </label>
      <input type="email" id="doc-vld-email-2" placeholder="Input email here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">URL: </label>
      <input type="url" id="doc-vld-url-2" placeholder="Input URL here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">Age: </label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="Between 18 and 120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">Favorite fruit: </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Orange" name="docVlCb" minchecked="2" maxchecked="4" required> Orange
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Apple" name="docVlCb"> Apple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Pineapple" name="docVlCb"> Pineapple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Mongo" name="docVlCb"> Mongo
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Banana" name="docVlCb"> Banana
      </label>
    </div>

    <div class="am-form-group">
      <label>Gender: </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> Male
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> Female
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">Drowdown</label>
      <select id="doc-select-1" required>
        <option value="">-=Please select one=-</option>
        <option value="option1">Option 1...</option>
        <option value="option2">Option 2.....</option>
        <option value="option3">Option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">Checkbox</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">Comment: </label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>JS Form Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username: </label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">Email: </label>
      <input type="email" id="doc-vld-email-2" placeholder="Input email here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">URL: </label>
      <input type="url" id="doc-vld-url-2" placeholder="Input URL here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">Age: </label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="Between 18 and 120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">Favorite fruit: </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Orange" name="docVlCb" minchecked="2" maxchecked="4" required> Orange
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Apple" name="docVlCb"> Apple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Pineapple" name="docVlCb"> Pineapple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Mongo" name="docVlCb"> Mongo
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Banana" name="docVlCb"> Banana
      </label>
    </div>

    <div class="am-form-group">
      <label>Gender: </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> Male
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> Female
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">Drowdown</label>
      <select id="doc-select-1" required>
        <option value="">-=Please select one=-</option>
        <option value="option1">Option 1...</option>
        <option value="option2">Option 2.....</option>
        <option value="option3">Option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">Checkbox</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">Comment: </label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
```

### Show Prompts

Prompts can be shown using the callback interface of `.onValid` and `onInValid`.

See following examples for more details.

**Attention:** `.getValidationMessage(validity)` is a new method in `v2.3`, the former version only have customized information.

#### Show prompt at the bottom

`````html
<form action="" class="am-form" id="doc-vld-msg">
  <fieldset>
    <legend>JS Form Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username: </label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">Email: </label>
      <input type="email" id="doc-vld-email-2" placeholder="Input email here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">URL: </label>
      <input type="url" id="doc-vld-url-2" placeholder="Input URL here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">Age: </label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="Between 18 and 120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">Favorite fruit: </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Orange" name="docVlCb" minchecked="2" maxchecked="4" required> Orange
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Apple" name="docVlCb"> Apple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Pineapple" name="docVlCb"> Pineapple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Mongo" name="docVlCb"> Mongo
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Banana" name="docVlCb"> Banana
      </label>
    </div>

    <div class="am-form-group">
      <label>Gender: </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> Male
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> Female
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">Drowdown</label>
      <select id="doc-select-1" required>
        <option value="">-=Please select one=-</option>
        <option value="option1">Option 1...</option>
        <option value="option2">Option 2.....</option>
        <option value="option3">Option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">Checkbox</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">Comment: </label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
<script>
  $(function() {
    $('#doc-vld-msg').validator({
      onValid: function(validity) {
        $(validity.field).closest('.am-form-group').find('.am-alert').hide();
      },

      onInValid: function(validity) {
        var $field = $(validity.field);
        var $group = $field.closest('.am-form-group');
        var $alert = $group.find('.am-alert');
        // Use customized prompts or default prompts.
        var msg = $field.data('validationMessage') || this.getValidationMessage(validity);

        if (!$alert.length) {
          $alert = $('<div class="am-alert am-alert-danger"></div>').hide().
            appendTo($group);
        }

        $alert.html(msg).show();
      }
    });
  });
</script>
`````

```html
<form action="" class="am-form" id="doc-vld-msg">
  <fieldset>
    <legend>JS Form Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username: </label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">Email: </label>
      <input type="email" id="doc-vld-email-2" placeholder="Input email here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">URL: </label>
      <input type="url" id="doc-vld-url-2" placeholder="Input URL here" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">Age: </label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="Between 18 and 120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">Favorite fruit: </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Orange" name="docVlCb" minchecked="2" maxchecked="4" required> Orange
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Apple" name="docVlCb"> Apple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Pineapple" name="docVlCb"> Pineapple
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Mongo" name="docVlCb"> Mongo
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="Banana" name="docVlCb"> Banana
      </label>
    </div>

    <div class="am-form-group">
      <label>Gender: </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> Male
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> Female
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">Drowdown</label>
      <select id="doc-select-1" required>
        <option value="">-=Please select one=-</option>
        <option value="option1">Option 1...</option>
        <option value="option2">Option 2.....</option>
        <option value="option3">Option 3........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">Checkbox</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">Comment: </label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
```

```js
$(function() {
  $('#doc-vld-msg').validator({
    onValid: function(validity) {
      $(validity.field).closest('.am-form-group').find('.am-alert').hide();
    },

    onInValid: function(validity) {
      var $field = $(validity.field);
      var $group = $field.closest('.am-form-group');
      var $alert = $group.find('.am-alert');
      // Use customized prompts or default prompts.
      var msg = $field.data('validationMessage') || this.getValidationMessage(validity);

      if (!$alert.length) {
        $alert = $('<div class="am-alert am-alert-danger"></div>').hide().
          appendTo($group);
      }

      $alert.html(msg).show();
    }
  });
});
```

#### Tooltip

`````html
<form action="" class="am-form" id="form-with-tooltip">
  <fieldset>
    <legend>Define Tooltip</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-0">Username: </label>
      <input type="text" id="doc-vld-name-2-0" minlength="3"
             placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1-0">Password: </label>
      <input type="password" id="doc-vld-pwd-1-0" placeholder="Input 6 digits as password" pattern="^\d{6}$" required data-foolish-msg="Give me your password of IQ card!"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>

<style>
  #vld-tooltip {
    position: absolute;
    z-index: 1000;
    padding: 5px 10px;
    background: #F37B1D;
    min-width: 150px;
    color: #fff;
    transition: all 0.15s;
    box-shadow: 0 0 5px rgba(0,0,0,.15);
    display: none;
  }

  #vld-tooltip:before {
    position: absolute;
    top: -8px;
    left: 50%;
    width: 0;
    height: 0;
    margin-left: -8px;
    content: "";
    border-width: 0 8px 8px;
    border-color: transparent transparent #F37B1D;
    border-style: none inset solid;
  }
</style>

<script>
$(function() {
  var $form = $('#form-with-tooltip');
  var $tooltip = $('<div id="vld-tooltip">Prompts</div>');
  $tooltip.appendTo(document.body);

  $form.validator();

  var validator = $form.data('amui.validator');

  $form.on('focusin focusout', '.am-form-error input', function(e) {
    if (e.type === 'focusin') {
      var $this = $(this);
      var offset = $this.offset();
      var msg = $this.data('foolishMsg') || validator.getValidationMessage($this.data('validity'));

      $tooltip.text(msg).show().css({
        left: offset.left + 10,
        top: offset.top + $(this).outerHeight() + 10
      });
    } else {
      $tooltip.hide();
    }
  });
});
</script>
`````

```html
<form action="" class="am-form" id="form-with-tooltip">
  <fieldset>
    <legend>Define Tooltip</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-0">Username: </label>
      <input type="text" id="doc-vld-name-2-0" minlength="3"
             placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1-0">Password: </label>
      <input type="password" id="doc-vld-pwd-1-0" placeholder="Input 6 digits as password" pattern="^\d{6}$" required data-foolish-msg="Give me your password of IQ card!"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>

<style>
  #vld-tooltip {
    position: absolute;
    z-index: 1000;
    padding: 5px 10px;
    background: #F37B1D;
    min-width: 150px;
    color: #fff;
    transition: all 0.15s;
    box-shadow: 0 0 5px rgba(0,0,0,.15);
    display: none;
  }

  #vld-tooltip:before {
    position: absolute;
    top: -8px;
    left: 50%;
    width: 0;
    height: 0;
    margin-left: -8px;
    content: "";
    border-width: 0 8px 8px;
    border-color: transparent transparent #F37B1D;
    border-style: none inset solid;
  }
</style>
```
```js
$(function() {
  var $form = $('#form-with-tooltip');
  var $tooltip = $('<div id="vld-tooltip">Prompts</div>');
  $tooltip.appendTo(document.body);

  $form.validator();

  var validator = $form.data('amui.validator');

  $form.on('focusin focusout', '.am-form-error input', function(e) {
    if (e.type === 'focusin') {
      var $this = $(this);
      var offset = $this.offset();
      var msg = $this.data('foolishMsg') || validator.getValidationMessage($this.data('validity'));

      $tooltip.text(msg).show().css({
        left: offset.left + 10,
        top: offset.top + $(this).outerHeight() + 10
      });
    } else {
      $tooltip.hide();
    }
  });
});
```


### Equivalent Validation

Use `data-equal-to` to specify the field that need to be validated.

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>Password Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username：</label>
      <input type="text" id="doc-vld-name-2" minlength="3"
             placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1">Password: </label>
      <input type="password" id="doc-vld-pwd-1" placeholder="Input 6-digit password" pattern="^\d{6}$" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-2">Confirm Password：</label>
      <input type="password" id="doc-vld-pwd-2" placeholder="Please input the same password as above" data-equal-to="#doc-vld-pwd-1" required/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>Password Validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">Username：</label>
      <input type="text" id="doc-vld-name-2" minlength="3"
             placeholder="At least 3 characters" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1">Password: </label>
      <input type="password" id="doc-vld-pwd-1" placeholder="Input 6-digit password" pattern="^\d{6}$" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-2">Confirm Password：</label>
      <input type="password" id="doc-vld-pwd-2" placeholder="Please input the same password as above" data-equal-to="#doc-vld-pwd-1" required/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
```

### Customized Validation

There are some requirement that build-in validations can't help with. Therefore, we allow users to customize validation using `validate` option.

```javascript
$('#your-form').validator({
  validate: function(validity) {
    // Your validation
  }
```

The parameter `validity` is an object like [H5 ValidityState](https://developer.mozilla.org/en-US/docs/Web/API/ValidityState) attribute. Normally we will only use:

- `validity.field` - DOM Object. The field that is being validated. Can be transfer to jQuery object through `$(validity.field)`;
- `validity.valid` - Boolean. Whether this field pass the validation. Set it to `true` if pass. Otherwise, set it to `false`.

Use following attributes to discribe the detail of validation error:

```javascript
{
  customError: false,
  patternMismatch: false,
  rangeOverflow: false, // higher than maximum
  rangeUnderflow: false, // lower than  minimum
  stepMismatch: false,
  tooLong: false,
  // value is not in the correct syntax
  typeMismatch: false,
  // Returns true if the element has no value but is a required field
  valueMissing: false
}
```

Three validation attribute extended by plugin and their corresponding error name are:

- `minlength` -> `tooShort`
- `minchecked` -> `checkedUnderflow`
- `maxchecked` -> `checkedOverflow`

The HTML5 validator use these error details to generate prompts. ~~插件中暂未使用到这些属性，如果实在不想写，可以略过，~~
The plug in also use these details to generate prompts since `v2.3`.

**Attention**

- Use `validity.valid` to return whether validation is passed;
- If this is a remote validation, a [Deferred Object](http://api.jquery.com/category/deferred-object/) must be returned, and validity must be returned by callback function.

```javascript
return $.ajax({
    url: '...',
    // cache: false. Disable cache in real application.
    dataType: 'json'
  }).then(function(data) {
    // Ajax request success. Set validity.valid = true or false accordingly.

    // return validity
    return validity;
  }, function() {
    // Ajax request fail. Set validity.valid = true or false and return validity.
    return validity;
  });
```

`````html
<form action="" class="am-form" id="doc-vld-ajax">
  <fieldset>
    <legend>Customized validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-ajax-count">Ajax Server-side Validation: </label>
      <input type="text" class="js-ajax-validate" id="doc-vld-ajax-count"
             placeholder="Enter 10" data-validate-async/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-sync">Client-side Validation: </label>
      <input type="text" class="js-sync-validate" id="doc-vld-sync"
             placeholder="Enter 10"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" id="doc-vld-ajax">
  <fieldset>
    <legend>Customized validation</legend>
    <div class="am-form-group">
      <label for="doc-vld-ajax-count">Ajax Server-side Validation: </label>
      <input type="text" class="js-ajax-validate" id="doc-vld-ajax-count"
             placeholder="Enter 10" data-validate-async/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-sync">Client-side Validation: </label>
      <input type="text" class="js-sync-validate" id="doc-vld-sync"
             placeholder="Enter 10"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
```
<script>
  $(function() {
    $('#doc-vld-ajax').validator({
      validate: function(validity) {
        var v = $(validity.field).val();

        var comparer = function(v1, v2) {
          if (v1 != v2) {
            validity.valid = false;
          }

          // These attributes are currently not used, so you may dismiss them.
          if (v2 < 10) {
            validity.rangeUnderflow = true;
          } else if(v2 > 10) {
            validity.rangeOverflow = true;
          }
        };

        // Ajax Validation
        if ($(validity.field).is('.js-ajax-validate')) {
          // Return Deferred object
          return $.ajax({
            url: 'http://s.amazeui.org/media/i/demos/validate.json',
            // cache: false    Disable cache in real application. 
            dataType: 'json'
          }).then(function(data) {
            comparer(data.count, v);
            return validity;
          }, function() {
            return validity;
          });
        }

        // Local valication. No return.
        if ($(validity.field).is('.js-sync-validate')) {
          comparer(10, v);
        }
      }
    });
  })
</script>

```javascript
$(function() {
  $('#doc-vld-ajax').validator({
    validate: function(validity) {
      var v = $(validity.field).val();

      var comparer = function(v1, v2) {
        if (v1 != v2) {
          validity.valid = false;
        }

        // These attributes are currently not used, so you may dismiss them.
        if (v2 < 10) {
          validity.rangeUnderflow = true;
        } else if(v2 > 10) {
          validity.rangeOverflow = true;
        }
      };

      // Ajax Validation
      if ($(validity.field).is('.js-ajax-validate')) {
        // Return Deferred object
        return $.ajax({
          url: 'http://s.amazeui.org/media/i/demos/validate.json',
          // cache: false    Disable cache in real application. 
          dataType: 'json'
        }).then(function(data) {
          comparer(data.count, v);
          return validity;
        }, function() {
          return validity;
        });
      }

      // Local valication. No return.
      if ($(validity.field).is('.js-sync-validate')) {
        comparer(10, v);
      }
    }
  });
})
```

## Usage

### Using Data API

Add `data-am-validator` attribute to `form`.

### Using JS

```javascript
$(function() {
  $('#your-form').validator(options);
});
```

#### Options

```javascript
{
  // Whether use H5 validation.
  H5validation: false,

  // Native H5 input types. Don't need to add paterns for them.
  H5inputType: ['email', 'url', 'number'],

  // Use regular expression as patterns
  // key1: /^...$/. Fields include `js-pattern-key1` will be validated with this regular expression
  patterns: {},

  // Prefix of rule classes
  patternClassPrefix: 'js-pattern-',

  activeClass: 'am-active',

  // The class added to field when it doesn't pass the validation
  inValidClass: 'am-field-error',

  // The class added to field when it pass the validation
  validClass: 'am-field-valid',

  // Validate on submit
  validateOnSubmit: true,

  // The field need to be validated on submit
  // Elements to validate with allValid (only validating visible elements)
  // :input: selects all input, textarea, select and button elements.
  allFields: ':input:visible:not(:button, :disabled, .am-novalidate)',

  // The event that call validate() method
  customEvents: 'validate',

  // The following events in following fields will call validation
  keyboardFields: ':input:not(:button, :disabled,.am-novalidate)',
  keyboardEvents: 'focusout, change', // keyup, focusin

  // Validate when keyup fired on the element marked as `.am-active`( invalid validation will be given this class) 
  activeKeyup: false,

  // Validate when keyup fired on textarea[maxlength] element
  textareaMaxlenthKeyup: true,

  // Click on following field will cause validation
  pointerFields: 'input[type="range"]:not(:disabled, .am-novalidate), ' +
  'input[type="radio"]:not(:disabled, .am-novalidate), ' +
  'input[type="checkbox"]:not(:disabled, .am-novalidate), ' +
  'select:not(:disabled, .am-novalidate), ' +
  'option:not(:disabled, .am-novalidate)',
  pointerEvents: 'click',

  // The valid callback function
  onValid: function(validity) {
  },

  // The invalid callback function
  onInValid: function(validity) {
  },

  // This function is called when a field pass the validation. Prompts can be shown in this funciton.
  markValid: function(validity) {
    // this is Validator instance
    var options = this.options;
    var $field  = $(validity.field);
    var $parent = $field.closest('.am-form-group');
    $field.addClass(options.validClass).
      removeClass(options.inValidClass);

    $parent.addClass('am-form-success').removeClass('am-form-error');

    options.onValid.call(this, validity);
  },

  // This function is called when a field fail the validation. Prompts can be shown in this funciton.
  markInValid: function(validity) {
    var options = this.options;
    var $field  = $(validity.field);
    var $parent = $field.closest('.am-form-group');
    $field.addClass(options.inValidClass + ' ' + options.activeClass).
      removeClass(options.validClass);

    $parent.addClass('am-form-error').removeClass('am-form-success');

    options.onInValid.call(this, validity);
  },

  // Customized validation
  validate: function(validity) {
    // return validity;
  },

  // Customized submit
  //   - If there is no defination and `validateOnSubmit` is `true`, the whole form will be validated when submission.
  //   - If submission is defined here, `validateOnSubmit` will be invalid
  //        function(e) {
  //          // Use this.isFormValid() to get the form validation status
  //          // Attention: If the customized validation is an async programe, this.isFormValid() will return a Promise instead of a boolean. 
  //          // Do something...
  //        }
  submit: null
}
```

**Async Validation**

When the validation is an async program, `isFormValid()` returns a Promise, which can use [`jQuery.when()`](http://api.jquery.com/jQuery.when/) to handle.

```js
$('#xx').validator({
  submit: function() {
    var formValidity = this.isFormValid();

    $.when(formValidity).then(function() {
      // Success callback
    }, function() {
      // Error callback
    });
  }
});
```

#### Extended Regular Expression

The following operations are done before DOM get ready: 

```javascript
(function($) {
  if ($.AMUI && $.AMUI.validator) {
    // Add multiple RegExps
    $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
      colorHex: /^(#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}))?$/
    });
    // Add single RegExp
    $.AMUI.validator.patterns.yourpattern = /^your$/;
  }
})(window.jQuery);
```

`````html
<form action="" class="am-form" data-am-validator>
  <div class="am-form-group">
    <label for="">Input a color</label>
    <input type="text" class="js-pattern-colorHex" placeholder="Must be #xxx or #xxxxxx"/>
  </div>
  <div class="am-form-group">
    <label for="">your pattern</label>
    <input type="text" class="js-pattern-yourpattern" placeholder="Required. Must be your" required/>
  </div>
  <div>
    <button class="am-btn am-btn-secondary">Submit</button>
  </div>
</form>
<script>
  (function($) {
    if ($.AMUI && $.AMUI.validator) {
      // Add multiple RegExps
      $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
        colorHex: /^(#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}))?$/
      });
      // Add single RegExp
      $.AMUI.validator.patterns.yourpattern = /^your$/;
    }
  })(window.jQuery);
</script>
`````
```html
<form action="" class="am-form" data-am-validator>
  <div class="am-form-group">
    <label for="">Input a color</label>
    <input type="text" class="js-pattern-colorHex" placeholder="Must be #xxx or #xxxxxx"/>
  </div>
  <div class="am-form-group">
    <label for="">your pattern</label>
    <input type="text" class="js-pattern-yourpattern" placeholder="Required. Must be your" required/>
  </div>
  <div>
    <button class="am-btn am-btn-secondary">Submit</button>
  </div>
</form>
```

## Issue Test

### [#528](https://github.com/allmobilize/amazeui/issues/528)

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>Issue 528</legend>
    <div class="am-form-group">
      <label for="doc-vld-528">Phone：</label>
      <input type="text" id="doc-vld-528" class="js-pattern-mobile"
             placeholder="Enter phone number" required/>
    </div>
    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
<script>
  if ($.AMUI && $.AMUI.validator) {
    $.AMUI.validator.patterns.mobile = /^\s*1\d{10}\s*$/;
  }
</script>
`````
```html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>Issue 528</legend>
    <div class="am-form-group">
      <label for="doc-vld-528">Phone：</label>
      <input type="text" id="doc-vld-528" class="js-pattern-mobile"
             placeholder="Enter phone number" required/>
    </div>
    <button class="am-btn am-btn-secondary" type="submit">Submit</button>
  </fieldset>
</form>
<script>
  if ($.AMUI && $.AMUI.validator) {
    $.AMUI.validator.patterns.mobile = /^\s*1\d{10}\s*$/;
  }
</script>
```

## Reference

### Frequently used RegExps:

- Phone number with zone number：`/^(\d{3}-|\d{4}-)(\d{8}|\d{7})$/`
- Chinese characters： `/[\u4e00-\u9fa5]/`
- Zip code in China： `/^\d{6}$/`
- Mobile phone number in China： `/^1((3|5|8){1}\d{1}|70)\d{8}$/` 

### Attention

- `checkbox`/`radio` **These two element can't work correctly without `name` attribute**;
- `<input type="number">` returns an empty string `""` when input is not an number；
- Some browser select the first option in dropdown by default. Set the first option to be empty by using `value=""`.

```html
<div class="am-form-group">
  <label for="doc-select-1">Dropdown</label>
  <select id="doc-select-1" required>
    <option value="">-=Select one of the options=-</option>
    <option value="option1">Option 1...</option>
    <option value="option2">Option 2.....</option>
    <option value="option3">Option 3........</option>
  </select>
  <span class="am-form-caret"></span>
</div>
```

### Reference

- [Validity State](https://developer.mozilla.org/en-US/docs/Web/API/ValidityState)
- [HTML5 Constraint Validation API](http://dev.w3.org/html5/spec-preview/constraints.html#the-constraint-validation-api)
- [Constraint Validation: Native Client Side Validation for Web Forms](http://www.html5rocks.com/en/tutorials/forms/constraintvalidation/)
- https://github.com/wenzhixin/multiple-select/
- [HTML5 Placeholder jQuery Plugin](https://github.com/mathiasbynens/jquery-placeholder)
