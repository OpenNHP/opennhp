# Input group
------

Input group 基于 Form 组件和 Button 组件扩展，依赖这两个组件。

在容器上添加 `.am-input-group`，在标签文字上添加 `.am-input-group-label`，具体请查看示例代码。

## 基本演示

下面的代码中演示了结合 Icon 组件及添加文字的样式。

`````html
<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-user"></i></span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-lock"></i></span>
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
  <span class="am-input-group-label"><i class="am-icon-user"></i></span>
  <input type="text" class="am-form-field" placeholder="Username">
</div>

<div class="am-input-group">
  <span class="am-input-group-label"><i class="am-icon-lock"></i></span>
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

## 尺寸

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


## 复选框与单选框

将单选框与复选框放入 `.am-input-group-label` 内。

`````html
<div class="am-g">
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-col col-lg-6">
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
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="checkbox">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-label">
        <input type="radio">
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
</div>
```

## 结合 Button

需要用 `.am-input-group-btn` 包住按钮，而不是 `.am-input-group-label`。

`````html
<div class="am-g">
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">Go!</button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <input type="text" class="am-form-field">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">Search!</button>
      </span>
    </div>
  </div>
</div>
`````
```html
<div class="am-g">
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">Go!</button>
      </span>
      <input type="text" class="am-form-field">
    </div>
  </div>
  <div class="am-col col-lg-6">
    <div class="am-input-group">
      <input type="text" class="am-form-field">
      <span class="am-input-group-btn">
        <button class="am-btn am-btn-default" type="button">Search!</button>
      </span>
    </div>
  </div>
</div>
```