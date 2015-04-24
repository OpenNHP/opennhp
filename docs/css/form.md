# Form 表单元素
---

`<form>` 元素样式。

## 基本使用


### 单选、复选框

`checkbox`、`radio` 类型的 `<input>` 与其他元素稍有区别：

- 块级显示时在容器上添加 `.am-checkbox`、`.am-radio` class；
- 行内显示时在容器上添加 `.am-checkbox-inline`、`.am-radio-inline` class。

### 下拉选框

**`<select>` 是一个比较奇葩的元素，长得丑还不让人给它打扮**。

单使用 CSS， 很难给 `select` 定义跨浏览器兼容的样式，保留浏览器默认样式可能是它最好的归宿（[Pure CSS 就是这么干的](http://purecss.io/forms/#stacked-form))。Amaze UI 中针对 Webkit 浏览器写了一点样式替换了默认的下上三角形。

### 文件选择域

`<input type="file">` 也是 CSS 啃不动的一块骨头，如果实在看不惯原生的样式，一般的做法是把文件选择域设为透明那个，覆盖在其他元素。

`````html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-default am-btn-sm">
    <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
  <input type="file" multiple>
</div>

<hr/>

<div class="am-form-group am-form-file">
  <i class="am-icon-cloud-upload"></i> 选择要上传的文件
  <input type="file" multiple>
</div>
`````

```html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-default am-btn-sm">
    <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
  <input type="file" multiple>
</div>

<hr/>

<div class="am-form-group am-form-file">
  <i class="am-icon-cloud-upload"></i> 选择要上传的文件
  <input type="file" multiple>
</div>
```

存在的问题是不会显示已经选择的文件，对用户不够友好，需要[配合 JS](https://developer.mozilla.org/en-US/docs/Web/API/FileList) 使用：

`````html
<div class="am-form-group am-form-file">
  <button type="button" class="am-btn am-btn-danger am-btn-sm">
    <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
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
    <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
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

### 基本演示

在容器上添加 `.am-form` class，容器里的子元素才会应用 Amaze UI 定义的样式。

`````html
<form class="am-form">
  <fieldset>
    <legend>表单标题</legend>

    <div class="am-form-group">
      <label for="doc-ipt-email-1">邮件</label>
      <input type="email" class="" id="doc-ipt-email-1" placeholder="输入电子邮件">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-pwd-1">密码</label>
      <input type="password" class="" id="doc-ipt-pwd-1" placeholder="设置个密码吧">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-file-1">原生文件上传域</label>
      <input type="file" id="doc-ipt-file-1">
      <p class="am-form-help">请选择要上传的文件...</p>
    </div>

    <div class="am-form-group am-form-file">
      <label for="doc-ipt-file-2">Amaze UI 文件上传域</label>
      <div>
        <button type="button" class="am-btn am-btn-default am-btn-sm">
          <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
      </div>
      <input type="file" id="doc-ipt-file-2">
    </div>

    <hr/>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> 复选框，选我选我选我
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option1" checked>
        单选框 - 选项1
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option2">
        单选框 - 选项2
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option1"> 选我
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option2"> 同时可以选我
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option3"> 还可以选我
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docInlineRadio"> 每一分
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> 每一秒
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> 多好
      </label>
    </div>

    <div class="am-form-group am-form-select">
      <label for="doc-select-1">下拉多选框</label>
      <select id="doc-select-1">
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">多选框</label>
      <select multiple class="" id="doc-select-2">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-ta-1">文本域</label>
      <textarea class="" rows="5" id="doc-ta-1"></textarea>
    </div>

    <p><button type="submit" class="am-btn am-btn-default">提交</button></p>
  </fieldset>
</form>
`````

```html
<form class="am-form">
  <fieldset>
    <legend>表单标题</legend>

    <div class="am-form-group">
      <label for="doc-ipt-email-1">邮件</label>
      <input type="email" class="" id="doc-ipt-email-1" placeholder="输入电子邮件">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-pwd-1">密码</label>
      <input type="password" class="" id="doc-ipt-pwd-1" placeholder="设置个密码吧">
    </div>

    <div class="am-form-group">
      <label for="doc-ipt-file-1">原生文件上传域</label>
      <input type="file" id="doc-ipt-file-1">
      <p class="am-form-help">请选择要上传的文件...</p>
    </div>

    <div class="am-form-group am-form-file">
      <label for="doc-ipt-file-2">Amaze UI 文件上传域</label>
      <div>
        <button type="button" class="am-btn am-btn-default am-btn-sm">
          <i class="am-icon-cloud-upload"></i> 选择要上传的文件</button>
      </div>
      <input type="file" id="doc-ipt-file-2">
    </div>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> 复选框，选我选我选我
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option1" checked>
        单选框 - 选项1
      </label>
    </div>

    <div class="am-radio">
      <label>
        <input type="radio" name="doc-radio-1" value="option2">
        单选框 - 选项2
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option1"> 选我
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option2"> 同时可以选我
      </label>
      <label class="am-checkbox-inline">
        <input type="checkbox" value="option3"> 还可以选我
      </label>
    </div>

    <div class="am-form-group">
      <label class="am-radio-inline">
        <input type="radio"  value="" name="docInlineRadio"> 每一分
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> 每一秒
      </label>
      <label class="am-radio-inline">
        <input type="radio" name="docInlineRadio"> 多好
      </label>
    </div>

    <div class="am-form-group">
      <label for="doc-select-1">下拉多选框</label>
      <select id="doc-select-1">
        <option value="option1">选项一...</option>
        <option value="option2">选项二.....</option>
        <option value="option3">选项三........</option>
      </select>
      <span class="am-form-caret"></span>
    </div>

    <div class="am-form-group">
      <label for="doc-select-2">多选框</label>
      <select multiple class="" id="doc-select-2">
        <option>1</option>
        <option>2</option>
        <option>3</option>
        <option>4</option>
        <option>5</option>
      </select>
    </div>

    <div class="am-form-group">
      <label for="doc-ta-1">文本域</label>
      <textarea class="" rows="5" id="doc-ta-1"></textarea>
    </div>

    <p><button type="submit" class="am-btn am-btn-default">提交</button></p>
  </fieldset>
</form>
```

### 表单形状

`````html
<p><input type="text" class="am-form-field am-radius" placeholder="圆角表单域" /></p>
<p><input type="text" class="am-form-field am-round" placeholder="椭圆表单域"/></p>
`````
```html
<p><input type="text" class="am-form-field am-radius" placeholder="圆角表单域" /></p>
<p><input type="text" class="am-form-field am-round" placeholder="椭圆表单域"/></p>
```

## 表单域状态

设置表单元素的不同状态。

### 禁用单个元素

给 `<input>` 添加 `disabled` 属性以禁用表单元素。

`````html
<form class="am-form">
  <input class="am-form-field" type="text" placeholder="禁止输入的表单..." disabled>
</form>
`````

```html
<form class="am-form">
  <input class="am-form-field" type="text" placeholder="禁止输入的表单..." disabled>
</form>
```

给 `<fieldset>` 元素增加 `disabled` 属性，禁用所有的子元素。

### 禁用域内的元素

`````html
<form class="am-form">
  <fieldset disabled>
    <div class="am-form-group">
      <label for="doc-ds-ipt-1">禁用的文本框</label>
      <input type="text" id="doc-ds-ipt-1" class="am-form-field" placeholder="禁止输入">
    </div>

    <div class="am-form-group">
      <label for="oc-ds-select-1">禁用的下拉选框</label>
      <select id="oc-ds-select-1" class="am-form-field">
        <option>禁止选择我</option>
      </select>
    </div>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> 无法选中的复选框
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
      <label for="doc-ds-ipt-1">禁用的文本框</label>
      <input type="text" id="doc-ds-ipt-1" class="am-form-field" placeholder="禁止输入">
    </div>

    <div class="am-form-group">
      <label for="oc-ds-select-1">禁用的下拉选框</label>
      <select id="oc-ds-select-1" class="am-form-field">
        <option>禁止选择我</option>
      </select>
    </div>

    <div class="am-checkbox">
      <label>
        <input type="checkbox"> 无法选中的复选框
      </label>
    </div>

    <button type="submit" class="am-btn am-btn-primary">Submit</button>
  </fieldset>
</form>
```

### 禁用链接

`<a>` 元素设置禁用状态需要加上 `.am-disabled`  class。

`````html
<a class="am-btn am-btn-primary am-disabled">禁止提交</a>
`````
```html
<a class="am-btn am-btn-primary am-disabled">禁止提交</a>
```

## 表单排列

### 水平排列

在 `<form>` 添加 `.am-form-horizontal` class 并结合网格系统使用。

`````html
<form class="am-form am-form-horizontal">
  <div class="am-form-group">
    <label for="doc-ipt-3" class="am-u-sm-2 am-form-label">电子邮件</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3" placeholder="输入你的电子邮件">
    </div>
  </div>

  <div class="am-form-group">
    <label for="doc-ipt-pwd-2" class="am-u-sm-2 am-form-label">密码</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-2" placeholder="设置一个密码吧">
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="checkbox">
        <label>
          <input type="checkbox"> 记住十万年
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">提交登入</button>
    </div>
  </div>
</form>
`````

```html
<form class="am-form am-form-horizontal">
  <div class="am-form-group">
    <label for="doc-ipt-3" class="am-u-sm-2 am-form-label">电子邮件</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3" placeholder="输入你的电子邮件">
    </div>
  </div>

  <div class="am-form-group">
    <label for="doc-ipt-pwd-2" class="am-u-sm-2 am-form-label">密码</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-2" placeholder="设置一个密码吧">
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="checkbox">
        <label>
          <input type="checkbox"> 记住十万年
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">提交登入</button>
    </div>
  </div>
</form>
```

### 行内排列

在外围容器上添加 `.am-form-inline`。 **注意**： 行内排列的元素并没有设置右边距，默认使用 `inline-block` 元素的间距，压缩 HTML 后行内表单元素的右边距会消失，需要自行处理。

`````html
<form class="am-form-inline" role="form">
  <div class="am-form-group">
    <input type="email" class="am-form-field" placeholder="电子邮件">
  </div>

  &nbsp;

  <div class="am-form-group">
    <input type="password" class="am-form-field" placeholder="密码">
  </div>

  &nbsp;

  <div class="am-checkbox">
    <label>
      <input type="checkbox"> 记住我
    </label>
  </div>

  &nbsp;

  <button type="submit" class="am-btn am-btn-default">登录</button>
</form>
`````
```html
<form class="am-form-inline" role="form">
  <div class="am-form-group">
    <input type="email" class="am-form-field" placeholder="电子邮件">
  </div>

  <div class="am-form-group">
    <input type="password" class="am-form-field" placeholder="密码">
  </div>

  <div class="am-checkbox">
    <label>
      <input type="checkbox"> 记住我
    </label>
  </div>

  <button type="submit" class="am-btn am-btn-default">登录</button>
</form>
```

## 表单域 Icon

表单 group 元素上添加 `.am-form-icon`，依赖 `icon` 组件。

`````html
<form action="" class="am-form am-form-inline">
  <div class="am-form-group am-form-icon">
    <i class="am-icon-calendar"></i>
    <input type="text" class="am-form-field" placeholder="日期">
  </div>
  &nbsp;
  <div class="am-form-group am-form-icon">
    <i class="am-icon-clock-o"></i>
    <input type="text" class="am-form-field" placeholder="时间">
  </div>
</form>
`````

```html
<form action="" class="am-form am-form-inline">
  <div class="am-form-group am-form-icon">
    <i class="am-icon-calendar"></i>
    <input type="text" class="am-form-field" placeholder="日期">
  </div>

  <div class="am-form-group am-form-icon">
    <i class="am-icon-clock-o"></i>
    <input type="text" class="am-form-field" placeholder="时间">
  </div>
</form>
```

## 验证状态

### 演示

`````html
<form action="" class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label class="am-form-label" for="doc-ipt-success">验证成功</label>
    <input type="text" id="doc-ipt-success" class="am-form-field">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning">
    <label class="am-form-label" for="doc-ipt-warning">验证警告</label>
    <input type="text" id="doc-ipt-warning" class="am-form-field">
  </div>
  <div class="am-form-group am-form-error">
    <label class="am-form-label" for="doc-ipt-error">验证失败</label>
    <input type="text" id="doc-ipt-error" class="am-form-field">
  </div>
</form>
`````

```html
<form action="" class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <label class="am-form-label" for="doc-ipt-success">验证成功</label>
    <input type="text" id="doc-ipt-success" class="am-form-field">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning">
    <label class="am-form-label" for="doc-ipt-warning">验证警告</label>
    <input type="text" id="doc-ipt-warning" class="am-form-field">
  </div>
  <div class="am-form-group am-form-error">
    <label class="am-form-label" for="doc-ipt-error">验证失败</label>
    <input type="text" id="doc-ipt-error" class="am-form-field">
  </div>
</form>
```
### 带图标的验证

添加 `.am-form-icon` 和 `.am-form-feedback`。

~~注意：Icon 的样式针对 `.am-form-group` 单行排列编写，多行的时候会出现位置不对的情况。~~

`v2.3.1` 中调整样式，支持带有 `label` 的情形。

`````html
<form class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证成功">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证警告">
    <span class="am-icon-warning"></span>
  </div>
  <div class="am-form-group am-form-error am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证失败">
    <span class="am-icon-times"></span>
  </div>
</form>
`````

```html
<form class="am-form">
  <div class="am-form-group am-form-success am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证成功">
    <span class="am-icon-check"></span>
  </div>
  <div class="am-form-group am-form-warning am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证警告">
    <span class="am-icon-warning"></span>
  </div>
  <div class="am-form-group am-form-error am-form-icon am-form-feedback">
    <input type="text" class="am-form-field" placeholder="验证失败">
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



## 表单域大小

### 单个域的大小

在表单元素上添加以下类名：

- `am-input-lg`
- `am-input-sm`

适用于没有 `<label>` 的表单，如果表单包含 `<label>` 且需要调整大小的，请往后看。

`````html
<form class="am-form">
  <input class="am-form-field am-input-lg" type="text" placeholder="添加了 .am-input-lg">
  <br/>
  <input class="am-form-field" type="text" placeholder="默认的 input">
  <br/>
  <input class="am-form-field am-input-sm" type="text" placeholder="添加了 .am-input-sm">
  <br/>

  <div class="am-form-group am-form-select">
    <select class=" am-input-lg">
      <option value="">添加了 .am-input-lg</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class="">
      <option value="">select 默认大小</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class=" am-input-sm">
      <option value="">添加了 .am-input-sm</option>
    </select>
  </div>
</form>
`````

```html
<form class="am-form">
  <input class="am-form-field am-input-lg" type="text" placeholder="添加了 .am-input-lg">
  <br/>
  <input class="am-form-field" type="text" placeholder="默认的 input">
  <br/>
  <input class="am-form-field am-input-sm" type="text" placeholder="添加了 .am-input-sm">
  <br/>

  <div class="am-form-group am-form-select">
    <select class=" am-input-lg">
      <option value="">添加了 .am-input-lg</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class="">
      <option value="">select 默认大小</option>
    </select>
  </div>

  <div class="am-form-group am-form-select">
    <select class=" am-input-sm">
      <option value="">添加了 .am-input-sm</option>
    </select>
  </div>
</form>
```

### 组大小

在 `.am-form-group` 的基础上添加以下 class，也可以实现对表单大小的设置：

- `.am-form-group-sm`
- `.am-form-group-lg`

注意**可输入类型的 `input` 元素上需要添加 `.am-form-field`**，不需要再添加 `.am-input-sm` 此类的 class。

`````html
<form class="am-form am-form-horizontal">
  <div class="am-form-group am-form-group-sm">
    <label for="doc-ipt-3-1" class="am-u-sm-2 am-form-label">电邮</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-1" class="am-form-field" placeholder="输入你的电子邮件">
    </div>
  </div>

  <div class="am-form-group am-form-group-lg">
    <label for="doc-ipt-pwd-21" class="am-u-sm-2 am-form-label">密码</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-21" class="am-form-field" placeholder="设置一个密码吧">
    </div>
  </div>

  <div class="am-form-group am-form-group-sm">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="am-checkbox">
        <label>
          <input type="checkbox"> 记住十万年
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">提交登入</button>
    </div>
  </div>
</form>
`````

```html
<form class="am-form am-form-horizontal">
  <!-- am-form-group 的基础上添加了 am-form-group-sm -->
  <div class="am-form-group am-form-group-sm">
    <label for="doc-ipt-3-1" class="am-u-sm-2 am-form-label">电邮</label>
    <div class="am-u-sm-10">
      <input type="email" id="doc-ipt-3-1" class="am-form-field" placeholder="输入你的电子邮件">
    </div>
  </div>

  <!-- am-form-group 的基础上添加了 am-form-group-lg -->
  <div class="am-form-group am-form-group-lg">
    <label for="doc-ipt-pwd-21" class="am-u-sm-2 am-form-label">密码</label>
    <div class="am-u-sm-10">
      <input type="password" id="doc-ipt-pwd-21" class="am-form-field" placeholder="设置一个密码吧">
    </div>
  </div>

  <div class="am-form-group am-form-group-sm">
    <div class="am-u-sm-offset-2 am-u-sm-10">
      <div class="am-checkbox">
        <label>
          <input type="checkbox"> 记住十万年
        </label>
      </div>
    </div>
  </div>

  <div class="am-form-group">
    <div class="am-u-sm-10 am-u-sm-offset-2">
      <button type="submit" class="am-btn am-btn-default">提交登入</button>
    </div>
  </div>
</form>
```

## 输入框组

使用 `.am-form-set` 嵌套一系列 `<input>` 元素。

`````html
<div class="am-g">
  <div class="am-u-md-8 am-u-sm-centered">
    <form class="am-form">
      <fieldset class="am-form-set">
        <input type="text" placeholder="取个名字">
        <input type="text" placeholder="设个密码">
        <input type="email" placeholder="填下邮箱">
      </fieldset>
      <button type="submit" class="am-btn am-btn-primary am-btn-block">注册个账号</button>
    </form>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-md-8 am-u-sm-centered">
    <form class="am-form">
      <fieldset class="am-form-set">
        <input type="text" placeholder="取个名字">
        <input type="text" placeholder="设个密码">
        <input type="email" placeholder="填下邮箱">
      </fieldset>
      <button type="submit" class="am-btn am-btn-primary am-btn-block">注册个账号</button>
    </form>
  </div>
</div>
```
## 参考链接

- [Webkit 浏览器 Radio/Checkbox 纯 CSS 样式](http://jsbin.com/gitovovidu)
