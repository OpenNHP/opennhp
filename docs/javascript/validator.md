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

基于 HTML5 的表单验证，使用 H5 `type`、`required`、`pattern`、`min`、`max`、`minlength`、`maxlength` 等属性进行验证，在不支持 JS 的环境中可以平稳退化到 H5 原生验证。

## 使用示例

### HTML5 原生表单验证

如果表单只面向 H5 浏览器，而且不需要过多的控制，那原生的表单验证无疑是省时省力的选择，通过 `:valid`、`:invalid` 伪类可以控制不同验证状态的样式。

`````html
<form action="" class="am-form">
  <fieldset>
    <legend>H5 原生表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-1">用户名：</label>
      <input type="text" id="doc-vld-name-1" maxlength="3" pattern="^\d+$" placeholder="输入用户名" class="am-form-field" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-1">邮箱：</label>
      <input type="email" id="doc-vld-email-1" placeholder="输入邮箱" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-1">网址：</label>
      <input type="url" id="doc-vld-url-1" placeholder="输入网址" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-1">年龄：</label>
      <input type="number" class=""  id="doc-vld-age-1" max="100" placeholder="输入年龄" required />
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-1">评论：</label>
      <textarea id="doc-vld-ta-1" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form">
  <fieldset>
    <legend>H5 原生表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name">用户名：</label>
      <input type="text" id="doc-vld-name" minlength="3" placeholder="输入用户名" class="am-form-field" required/>
    </div>
  </fieldset>
</form>
```

**参考链接：**

- [Forms in HTML](https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/Forms_in_HTML)
- [:invalid 伪类](https://developer.mozilla.org/en-US/docs/Web/CSS/:invalid)
- [HTML5 Form Validation](http://www.sitepoint.com/html5-form-validation/)
- [HTML5 Form Validation Examples](http://www.the-art-of-web.com/html/html5-form-validation/)

### JS 表单验证

JS 表单验证基于 HTML5 的各项验证属性进行：

- `required`: 必填；
- `pattern`: 验证正则表达式，插件内置了 `email`、`url`、`number` 三种类型的正则表达式；
- `minlength`/`maxlength`: 字符限制；
- `min`/`max`: 最小、最大值限制，仅适用于数值类型的域；
- `minchecked`/`maxchecked`: 至少、至多选择数，适用于 `checkbox`、下拉多选框，`checkbox` 时将相关属性的设置在同组的第一个元素上；
- `.js-pattern-xx`: 验证规则 class，正则库中存在的规则可以通过添加相应 class 实现规则添加。

**注意：**

HTML5 原生表单验证中 `pattern` 只验证值的合法性，也就是**可以不填，如果填写则必须符合规则**。如果是必填项，仍要添加 `required` 属性。该插件与 HTML5 的规则保持一致。

```html
<!-- 下面三种写法等效 -->
<!-- 只内置了 email url number 三种类型的正则，可自行扩展 -->
<input type="email"/>

<!-- js-pattern-xx 其中 xx 为 pattern 库中的 key -->
<input type="text" class="js-pattern-email"/>

<input type="text" pattern="^(...email regex...)$"/>
```

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>JS 表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">邮箱：</label>
      <input type="email" id="doc-vld-email-2" placeholder="输入邮箱" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">网址：</label>
      <input type="url" id="doc-vld-url-2" placeholder="输入网址" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">年龄：</label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="输入年龄  18-120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">爱好：</label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="橘子" name="docVlCb" minchecked="2" maxchecked="4" required> 橘子
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="苹果" name="docVlCb"> 苹果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="菠萝" name="docVlCb"> 菠萝
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="芒果" name="docVlCb"> 芒果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="香蕉" name="docVlCb"> 香蕉
      </label>
    </div>

    <div class="am-form-group">
      <label>性别： </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> 男
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 女
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 其他
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">下拉单选框</label>
      <select id="doc-select-1" required>
        <option value="">-=请选择一项=-</option>
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">多选框</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">评论：</label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>JS 表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2">邮箱：</label>
      <input type="email" id="doc-vld-email-2" placeholder="输入邮箱" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2">网址：</label>
      <input type="url" id="doc-vld-url-2" placeholder="输入网址" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2">年龄：</label>
      <input type="number" class=""  id="doc-vld-age-2" placeholder="输入年龄  18-120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">爱好：</label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="橘子" name="docVlCb" minchecked="2" maxchecked="4" required> 橘子
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="苹果" name="docVlCb"> 苹果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="菠萝" name="docVlCb"> 菠萝
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="芒果" name="docVlCb"> 芒果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="香蕉" name="docVlCb"> 香蕉
      </label>
    </div>

    <div class="am-form-group">
      <label>性别： </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> 男
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 女
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 其他
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">下拉单选框</label>
      <select id="doc-select-1" required>
        <option value="">-=请选择一项=-</option>
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">多选框</label>
      <select multiple class="" id="doc-select-2" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2">评论：</label>
      <textarea id="doc-vld-ta-2" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
```

### 显示提示信息

通过插件的 `.onValid` 和 `onInValid` 回调接口，可以根据需要定提示信息显示。

使用时可以自行定义提示信息，也可以使用插件的内置的提示信息，详见后面的示例代码。

**注意：** `.getValidationMessage(validity)` 是 `v2.3` 中新增的方法，以前的版本只能使用自定义信息。

#### 底部显示提示信息

`````html
<form action="" class="am-form" id="doc-vld-msg">
  <fieldset>
    <legend>JS 表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-1">用户名：</label>
      <input type="text" id="doc-vld-name-2-1" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2-1">邮箱：</label>
      <input type="email" id="doc-vld-email-2-1" data-validation-message="自定义提示信息：输入地球上的电子邮箱撒" placeholder="输入邮箱" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2-1">网址：</label>
      <input type="url" id="doc-vld-url-2-1" placeholder="输入网址" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2-1">年龄：</label>
      <input type="number" class=""  id="doc-vld-age-2-1" placeholder="输入年龄  18-120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">爱好：</label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="橘子" name="docVlCb" minchecked="2" maxchecked="4" required> 橘子
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="苹果" name="docVlCb"> 苹果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="菠萝" name="docVlCb"> 菠萝
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="芒果" name="docVlCb"> 芒果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="香蕉" name="docVlCb"> 香蕉
      </label>
    </div>

    <div class="am-form-group">
      <label>性别： </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> 男
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 女
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 其他
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1-1">下拉单选框</label>
      <select id="doc-select-1-1" required>
        <option value="">-=请选择一项=-</option>
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2-1">多选框</label>
      <select multiple class="" id="doc-select-2-1" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2-1">评论：</label>
      <textarea id="doc-vld-ta-2-1" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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
        // 使用自定义的提示信息 或 插件内置的提示信息
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
    <legend>显示验证提示信息</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-1">用户名：</label>
      <input type="text" id="doc-vld-name-2-1" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-email-2-1">邮箱：</label>
      <input type="email" id="doc-vld-email-2-1" data-validation-message="自定义提示信息：输入地球上的电子邮箱撒" placeholder="输入邮箱" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-url-2-1">网址：</label>
      <input type="url" id="doc-vld-url-2-1" placeholder="输入网址" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-age-2-1">年龄：</label>
      <input type="number" class=""  id="doc-vld-age-2-1" placeholder="输入年龄  18-120" min="18" max="120" required />
    </div>

    <div class="am-form-group">
      <label class="am-form-label">爱好：</label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="橘子" name="docVlCb" minchecked="2" maxchecked="4" required> 橘子
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="苹果" name="docVlCb"> 苹果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="菠萝" name="docVlCb"> 菠萝
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="芒果" name="docVlCb"> 芒果
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="香蕉" name="docVlCb"> 香蕉
      </label>
    </div>

    <div class="am-form-group">
      <label>性别： </label>
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docVlGender" required> 男
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 女
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docVlGender"> 其他
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1-1">下拉单选框</label>
      <select id="doc-select-1-1" required>
        <option value="">-=请选择一项=-</option>
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2-1">多选框</label>
      <select multiple class="" id="doc-select-2-1" minchecked="2" maxchecked="4">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-ta-2-1">评论：</label>
      <textarea id="doc-vld-ta-2-1" minlength="10" maxlength="100"></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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
      // 使用自定义的提示信息 或 插件内置的提示信息
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
    <legend>定义 Tooltip</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-0">用户名：</label>
      <input type="text" id="doc-vld-name-2-0" minlength="3"
             placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1-0">密码：</label>
      <input type="password" id="doc-vld-pwd-1-0" placeholder="6 位数字的银行卡密码" pattern="^\d{6}$" required data-foolish-msg="把 IQ 卡密码交出来！"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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
  var $tooltip = $('<div id="vld-tooltip">提示信息！</div>');
  $tooltip.appendTo(document.body);

  $form.validator({
    onValid: function() {
      $tooltip.hide();
    }
  });

  $form.on('focusin focusout', '.am-form-error input', function(e) {
    if (e.type === 'focusin') {
      var $this = $(this);
      var offset = $this.offset();
      var msg = $this.data('foolishMsg') || $form.validator('getValidationMessage', $this.data('validity'));

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
    <legend>定义 Tooltip</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2-0">用户名：</label>
      <input type="text" id="doc-vld-name-2-0" minlength="3"
             placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1-0">密码：</label>
      <input type="password" id="doc-vld-pwd-1-0" placeholder="6 位数字的银行卡密码" pattern="^\d{6}$" required data-foolish-msg="把 IQ 卡密码交出来！"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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
  var $tooltip = $('<div id="vld-tooltip">提示信息！</div>');
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


### 等值验证

通过 `data-equal-to` 指定要比较的域。

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>密码验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3"
             placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1">密码：</label>
      <input type="password" id="doc-vld-pwd-1" placeholder="6 位数字的银行卡密码" pattern="^\d{6}$" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-2">确认密码：</label>
      <input type="password" id="doc-vld-pwd-2" placeholder="请与上面输入的值一致" data-equal-to="#doc-vld-pwd-1" required/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>密码验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3"
             placeholder="输入用户名（至少 3 个字符）" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-1">密码：</label>
      <input type="password" id="doc-vld-pwd-1" placeholder="6 位数字的银行卡密码" pattern="^\d{6}$" required/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-pwd-2">确认密码：</label>
      <input type="password" id="doc-vld-pwd-2" placeholder="请与上面输入的值一致" data-equal-to="#doc-vld-pwd-1" required/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
```

### 自定义验证

插件预置的功能不可能满足各异的需求，通过 `validate` 选项，可以自定义验证规则，如远程验证等。

```javascript
$('#your-form').validator({
  validate: function(validity) {
    // 在这里编写你的验证逻辑
  }
```

参数 `validity` 是一个类似 [H5 ValidityState](https://developer.mozilla.org/en-US/docs/Web/API/ValidityState) 属性的对象。只要中主要用到的包括：

- `validity.field` - DOM 对象，当前验证的域，通过 `$(validity.field)` 可转换为 jQuery 对象，一般用于获取值和判断是否为特定域，以编写验证逻辑；
- `validity.valid` - 布尔值，验证是否通过，通过赋值 `true`，否则赋值 `false`。

其它属性用来描述验证出错的细节，包括：

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

插件扩展的三种验证属性，对应的自定义错误名称为：

- `minlength` -> `tooShort`
- `minchecked` -> `checkedUnderflow`
- `maxchecked` -> `checkedOverflow`

H5 浏览器原生验证通过错误细节来显示提示信息，~~插件中暂未使用到这些属性，如果实在不想写，可以略过，~~
`v2.3` 开始这些信息用于生成错误提示信息。

**需要注意的注意细节：**

- 通过 `validity.valid` 标记验证是否通过；
- 如果是远程异步验证，**必须**返回 [Deferred 对象](http://api.jquery.com/category/deferred-object/)，且回调函数中要返回 `validity`。

```javascript
return $.ajax({
    url: '...',
    // cache: false, 实际使用中请禁用缓存
    dataType: 'json'
  }).then(function(data) {
    // Ajax 请求成功，根据服务器返回的信息，设置 validity.valid = true or flase

    // 返回 validity
    return validity;
  }, function() {
    // Ajax 请求失败，根据需要决定验证是否通过，然后返回 validity
    return validity;
  });
```

`````html
<form action="" class="am-form" id="doc-vld-ajax">
  <fieldset>
    <legend>自定义验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-ajax-count">Ajax 服务器端验证：</label>
      <input type="text" class="js-ajax-validate" id="doc-vld-ajax-count"
             placeholder="只能填写数字 10" data-validate-async/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-sync">客户端验证：</label>
      <input type="text" class="js-sync-validate" id="doc-vld-sync"
             placeholder="只能填写数字 10"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
`````
```html
<form action="" class="am-form" id="doc-vld-ajax">
  <fieldset>
    <legend>自定义验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-ajax-count">Ajax 服务器端验证：</label>
      <input type="text" class="js-ajax-validate" id="doc-vld-ajax-count"
             placeholder="只能填写数字 10" data-validate-async/>
    </div>

    <div class="am-form-group">
      <label for="doc-vld-sync">客户端验证：</label>
      <input type="text" class="js-sync-validate" id="doc-vld-sync"
             placeholder="只能填写数字 10"/>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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

          // 这些属性目前没什么用，如果不想写可以忽略
          if (v2 < 10) {
            validity.rangeUnderflow = true;
          } else if(v2 > 10) {
            validity.rangeOverflow = true;
          }
        };

        // Ajax 验证
        if ($(validity.field).is('.js-ajax-validate')) {
          // 异步操作必须返回 Deferred 对象
          return $.ajax({
            url: 'http://s.amazeui.org/media/i/demos/validate.json',
            // cache: false, 实际使用中请禁用缓存
            dataType: 'json'
          }).then(function(data) {
            comparer(data.count, v);
            return validity;
          }, function() {
            return validity;
          });
        }

        // 本地验证，同步操作，无需返回值
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

        // 这些属性目前 v2.3 以前没什么用，如果不想写可以忽略
        // 从 v2.3 开始，这些属性被 getValidationMessage() 用于生成错误提示信息
        if (v2 < 10) {
          validity.rangeUnderflow = true;
        } else if(v2 > 10) {
          validity.rangeOverflow = true;
        }
      };

      // Ajax 验证
      if ($(validity.field).is('.js-ajax-validate')) {
        // 异步操作必须返回 Deferred 对象
        return $.ajax({
          url: 'http://s.amazeui.org/media/i/demos/validate.json',
          // cache: false, 实际使用中请禁用缓存
          dataType: 'json'
        }).then(function(data) {
          comparer(data.count, v);
          return validity;
        }, function() {
          return validity;
        });
      }

      // 本地验证，同步操作，无需返回值
      if ($(validity.field).is('.js-sync-validate')) {
        comparer(10, v);
        // return validity;
      }
    }
  });
});
```

### 验证 UEditor

Validator 可以和 [UEditor](http://ueditor.baidu.com/) 富文本编辑器结合使用。

`````html
<form action="" class="am-form" id="ue-form">
  <fieldset>
    <legend>JS 表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>
    <div class="am-form-group">
      <label for="doc-vld-ta-2">评论：</label>
      <textarea class="am-validate" name="myue" id="myue" minlength="10" maxlength="100" required></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
<script src="http://ueditor.baidu.com/ueditor/ueditor.config.js"></script>
<script src="http://ueditor.baidu.com/ueditor/ueditor.all.js"></script>
<script>
  $(function() {
    var $textArea = $('[name=myue');
    var editor = UE.getEditor('myue');
    var $form = $('#ue-form');

    $form.validator({
      submit: function() {
        // 同步编辑器数据
        editor.sync();

        var formValidity = this.isFormValid();

        // 表单验证未成功，而且未成功的第一个元素为 UEEditor 时，focus 编辑器
        if (!formValidity && $form.find('.' + this.options.inValidClass).eq(0).is($textArea)) {
          editor.focus();
        }

        console.warn('验证状态：', formValidity ? '通过' : '未通过');

        return false;
      }
    });

    // 编辑器内容变化时同步到 textarea
    editor.addListener('contentChange', function() {
      editor.sync();

      // 触发验证
      $('[name=myue]').trigger('change');
    });
  });
</script>
`````

```html
<form action="" class="am-form" id="ue-form">
  <fieldset>
    <legend>JS 表单验证</legend>
    <div class="am-form-group">
      <label for="doc-vld-name-2">用户名：</label>
      <input type="text" id="doc-vld-name-2" minlength="3" placeholder="输入用户名（至少 3 个字符）" required/>
    </div>
    <div class="am-form-group">
      <label for="doc-vld-ta-2">评论：</label>
      <textarea class="am-validate" name="myue" id="myue" required></textarea>
    </div>

    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
<script src="http://ueditor.baidu.com/ueditor/ueditor.config.js"></script>
<script src="http://ueditor.baidu.com/ueditor/ueditor.all.js"></script>
```

```javascript
$(function() {
  var $textArea = $('[name=myue');
  var editor = UE.getEditor('myue');
  var $form = $('#ue-form');

  $form.validator({
    submit: function() {
      // 同步编辑器数据
      editor.sync();

      var formValidity = this.isFormValid();

      // 表单验证未成功，而且未成功的第一个元素为 UEEditor 时，focus 编辑器
      if (!formValidity && $form.find('.' + this.options.inValidClass).eq(0).is($textArea)) {
        editor.focus();
      }

      console.warn('验证状态：', formValidity ? '通过' : '未通过');

      return false;
    }
  });

  // 编辑器内容变化时同步到 textarea
  editor.addListener('contentChange', function() {
    editor.sync();

    // 触发验证
    $('[name=myue]').trigger('change');
  });
});
```

## 使用方式

### 通过 Data API

在 `form` 上添加 `data-am-validator` 属性（同时可以设置相关选项）。

### 通过 JS

```javascript
$(function() {
  $('#your-form').validator(options);
});
```

#### 参数说明

```javascript
{
  // 是否使用 H5 原生表单验证，不支持浏览器会自动退化到 JS 验证
  H5validation: false,

  // 内置规则的 H5 input type，这些 type 无需添加 pattern
  H5inputType: ['email', 'url', 'number'],

  // 验证正则
  // key1: /^...$/，包含 `js-pattern-key1` 的域会自动应用改正则
  patterns: {},

  // 规则 class 钩子前缀
  patternClassPrefix: 'js-pattern-',

  activeClass: 'am-active',

  // 验证不通过时添加到域上的 class
  inValidClass: 'am-field-error',

  // 验证通过时添加到域上的 class
  validClass: 'am-field-valid',

  // 表单提交的时候验证
  validateOnSubmit: true,

  // 表单提交时验证的域
  // Elements to validate with allValid (only validating visible elements)
  // :input: selects all input, textarea, select and button elements.
  // @since 2.5: move `:visible` to `ignore` option, became to `:hidden`
  allFields: ':input:not(:button, :disabled, .am-novalidate)',

  // 表单提交时验证的忽略的域
  // ignored elements
  // @since 2.5
  ignore: ':hidden:not([data-am-selected], .am-validate)',

  // 调用 validate() 方法的自定义事件
  customEvents: 'validate',

  // 下列元素触发以下事件时会调用验证程序
  keyboardFields: ':input:not(:button, :disabled,.am-novalidate)',
  keyboardEvents: 'focusout, change', // keyup, focusin

  // 标记为 `.am-active` (发生错误以后添加此 class)的元素 keyup 时验证
  activeKeyup: false,

  // textarea[maxlength] 的元素 keyup 时验证
  textareaMaxlenthKeyup: true,

  // 鼠标点击下列元素时会调用验证程序
  pointerFields: 'input[type="range"]:not(:disabled, .am-novalidate), ' +
  'input[type="radio"]:not(:disabled, .am-novalidate), ' +
  'input[type="checkbox"]:not(:disabled, .am-novalidate), ' +
  'select:not(:disabled, .am-novalidate), ' +
  'option:not(:disabled, .am-novalidate)',
  pointerEvents: 'click',

  // 域通过验证时回调
  onValid: function(validity) {
  },

  // 验证出错时的回调， validity 对象包含相关信息，格式通 H5 表单元素的 validity 属性
  onInValid: function(validity) {
  },

  // 域验证通过时添加的操作，通过该接口可定义各种验证提示
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

  // 域验证失败时添加的操作，通过该接口可定义各种验证提示
  markInValid: function(validity) {
    var options = this.options;
    var $field  = $(validity.field);
    var $parent = $field.closest('.am-form-group');
    $field.addClass(options.inValidClass + ' ' + options.activeClass).
      removeClass(options.validClass);

    $parent.addClass('am-form-error').removeClass('am-form-success');

    options.onInValid.call(this, validity);
  },

  // 自定义验证程序接口，详见示例
  validate: function(validity) {
    // return validity;
  },

  // 定义表单提交处理程序
  //   - 如果没有定义且 `validateOnSubmit` 为 `true` 时，提交时会验证整个表单
  //   - 如果定义了表单提交处理程序，`validateOnSubmit` 将会失效
  //        function(e) {
  //          // 通过 this.isFormValid() 获取表单验证状态
  //          // 注意： 如果自定义验证程序而且自定义验证程序中包含异步验证的话 this.isFormValid() 返回的是 Promise，不是布尔值
  //          // Do something...
  //        }
  submit: null
}
```

**包含异步验证的表单提交处理**

包含异步验证时，`isFormValid()` 返回 Promise，可以使用 [`jQuery.when()`](http://api.jquery.com/jQuery.when/) 来处理结果。

```js
$('#xx').validator({
  submit: function() {
    var formValidity = this.isFormValid();

    $.when(formValidity).then(function() {
      // 验证成功的逻辑
    }, function() {
      // 验证失败的逻辑
    });
  }
});
```

#### 扩展正则库

在 DOM Ready 之前执行以下操作：

```javascript
(function($) {
  if ($.AMUI && $.AMUI.validator) {
    // 增加多个正则
    $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
      colorHex: /^(#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}))?$/
    });
    // 增加单个正则
    $.AMUI.validator.patterns.yourpattern = /^your$/;
  }
})(window.jQuery);
```

`````html
<form action="" class="am-form" data-am-validator>
  <div class="am-form-group">
    <label for="">输入一个颜色值</label>
    <input type="text" class="js-pattern-colorHex" placeholder="如果填写，必须是 #xxx 或 #xxxxxx"/>
  </div>
  <div class="am-form-group">
    <label for="">your pattern</label>
    <input type="text" class="js-pattern-yourpattern" placeholder="必填，且只能填 your" required/>
  </div>
  <div>
    <button class="am-btn am-btn-secondary">提交</button>
  </div>
</form>
<script>
  (function($) {
    if ($.AMUI && $.AMUI.validator) {
      // 增加多个正则
      $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
        colorHex: /^(#([a-fA-F0-9]{6}|[a-fA-F0-9]{3}))?$/
      });
      // 增加单个正则
      $.AMUI.validator.patterns.yourpattern = /^your$/;
    }
  })(window.jQuery);
</script>
`````
```html
<form action="" class="am-form" data-am-validator>
  <div class="am-form-group">
    <label for="">输入一个颜色值</label>
    <input type="text" class="js-pattern-colorHex" placeholder="如果填写，必须是 #xxx 或 #xxxxxx"/>
  </div>
  <div class="am-form-group">
    <label for="">your pattern</label>
    <input type="text" class="js-pattern-yourpattern" placeholder="必填，且只能填 your" required/>
  </div>
  <div>
    <button class="am-btn am-btn-secondary">提交</button>
  </div>
</form>
```

### 方法

- `.validator(options)` - 初始化表单验证
- `.validator('isFormValid')` - 返回表单验证状态，如果包含异步验证则返回 Promise（使用 `jQuery.when` 处理），否则返回布尔值

  ```js
  // 处理异步验证结果
  $.when($('myForm').validator('isFormValid')).then(function() {
    // 验证成功的逻辑
  }, function() {
    // 验证失败的逻辑
  });
  ```
- `.validator('destroy')` - 销毁表单验证

## Issue 测试

### [#528](https://github.com/allmobilize/amazeui/issues/528)

`````html
<form action="" class="am-form" data-am-validator>
  <fieldset>
    <legend>Issue 528</legend>
    <div class="am-form-group">
      <label for="doc-vld-528">手机号：</label>
      <input type="text" id="doc-vld-528" class="js-pattern-mobile"
             placeholder="输入手机号" required/>
    </div>
    <button class="am-btn am-btn-secondary" type="submit">提交</button>
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
      <label for="doc-vld-528">手机号：</label>
      <input type="text" id="doc-vld-528" class="js-pattern-mobile"
             placeholder="输入手机号" required/>
    </div>
    <button class="am-btn am-btn-secondary" type="submit">提交</button>
  </fieldset>
</form>
<script>
  if ($.AMUI && $.AMUI.validator) {
    $.AMUI.validator.patterns.mobile = /^\s*1\d{10}\s*$/;
  }
</script>
```

## 参考资源

### 常用正则表达式

- 带区号的电话号码：`/^(\d{3}-|\d{4}-)(\d{8}|\d{7})$/`
- 匹配中文字符： `/[\u4e00-\u9fa5]/`
- 国内邮政编码： `/^\d{6}$/`
- 国内手机号码： `/^1((3|5|8){1}\d{1}|70)\d{8}$/` （匹配 13x/15x/18x/170 号段，如有遗漏请自行添加）

### 注意事项

- `checkbox`/`radio` **务必添加 `name` 属性，否则无法正常工作**；
- `<input type="number">` 输入非数字字符时返回值为空字符串 `""`；
- 浏览器默认选中下拉单选框的第一项，使用时需将第一项的值设置为空 `value=""`。

```html
<div class="am-form-group">
  <label for="doc-select-1">下拉单选框</label>
  <select id="doc-select-1" required>
    <option value="">-=请选择一项=-</option>
    <option value="option1">选项一...</option>
    <option value="option2">选项二.....</option>
    <option value="option3">选项三........</option>
  </select>
  <span class="am-form-caret"></span>
</div>
```

### 参考链接

- [Validity State](https://developer.mozilla.org/en-US/docs/Web/API/ValidityState)
- [HTML5 Constraint Validation API](http://dev.w3.org/html5/spec-preview/constraints.html#the-constraint-validation-api)
- [Constraint Validation: Native Client Side Validation for Web Forms](http://www.html5rocks.com/en/tutorials/forms/constraintvalidation/)
- https://github.com/wenzhixin/multiple-select/
- [HTML5 Placeholder jQuery Plugin](https://github.com/mathiasbynens/jquery-placeholder)
