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
- `minlenth`/`maxlength`: 字符限制；
- `min`/`max`: 最小、最大值限制，仅适用于数值类型的域；
- `minchecked`/`maxchecked`: 至少、至多选择数，适用于 `checkbox`、下拉多选框，`checkbox` 时将相关属性的设置在同组的第一个元素上；
- `.js-pattern-xx`: 验证规则 class，正则库中存在的规则可以通过添加相应 class 实现规则添加。

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
  allFields: ':input:visible:not(:button, :disabled, .am-novalidate)',

  // 调用 validate() 方法的自定义事件
  customEvents: 'validate',

  // 下列元素触发以下事件时会调用验证程序
  keyboardFields: ':input:not(:button, :disabled,.am-novalidate)',
  keyboardEvents: 'focusout, change', // keyup, focusin

  activeKeyup: true,

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
  }
}
```

#### 扩展正则库

在 DOM Ready 之前执行以下操作：

```javascript
(function($) {
  if ($.AMUI && $.AMUI.validator) {
    // 增加多个正则
    $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
      colorHex: /^#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$/
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
  <div>
    <button class="am-btn am-btn-secondary">提交</button>
  </div>
</form>
<script>
  (function($) {
    if ($.AMUI && $.AMUI.validator) {
      // 增加多个正则
      $.AMUI.validator.patterns = $.extend($.AMUI.validator.patterns, {
        colorHex: /^#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$/
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
  <div>
    <button class="am-btn am-btn-secondary">提交</button>
  </div>
</form>
```

## 参考资源

### 常用正则表达式

- 带区号的电话号码：`/^(\d{3}-|\d{4}-)(\d{8}|\d{7})$/`
- 匹配中文字符： `/[\u4e00-\u9fa5]/`
- 国内邮政编码： `/^\d{6}$/`
- 国内手机号码： `/^1((3|5|8){1}\d{1}|70)\d{8}$/` （匹配 13x/15x/18x/170 号段，如有遗漏请自行添加）

### 注意事项

- `<input type="number">` 输入非数字字符时返回值为空字符串 `""`；

### 参考链接

- [Validity State](https://developer.mozilla.org/en-US/docs/Web/API/ValidityState)
- [HTML5 Constraint Validation API](http://dev.w3.org/html5/spec-preview/constraints.html#the-constraint-validation-api)
- [Constraint Validation: Native Client Side Validation for Web Forms](http://www.html5rocks.com/en/tutorials/forms/constraintvalidation/)
- https://github.com/wenzhixin/multiple-select/
- [HTML5 Placeholder jQuery Plugin](https://github.com/mathiasbynens/jquery-placeholder)
