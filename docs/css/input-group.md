# Input Group
-------------

Input group 基于 Form 组件和 Button 组件扩展，依赖这两个组件。

在容器上添加 `.am-input-group`，在标签文字上添加 `.am-input-group-label`，具体请查看示例代码。

## 基本使用

### 输入框与标签

下面的代码中演示了结合 Icon 组件及添加文字的样式。

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

### 复选/单选框与输入框

将单选框与复选框放入 `.am-input-group-label` 内。

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

### 输入框结合 Button

需要用 `.am-input-group-btn` 包住按钮，而不是 `.am-input-group-label`。

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

## 样式变换

### 尺寸

在 `.am-input-group` 添加标明尺寸的 class 即可。

包含 `.am-input-group-lg`、`.am-input-group-sm`。

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

### 颜色

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

使用按钮时除了在容器上设置颜色 class 外，还需要设置按钮的样式。

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
