# Datepicker
---

## 使用演示

### 基本形式

在 `input` 标签上增加 `data-am-datepicker` 属性，调用日期插件。

`````html
<p><input type="text" class="am-form-field am-radius" placeholder="日历组件" value="2014/12/07" data-am-datepicker /></p>

`````
```html
<p><input type="text" class="am-form-field am-radius" placeholder="日历组件" value="2014/12/07" data-am-datepicker /></p>
```

### 组件结合

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker>
  <input type="text" class="am-form-field" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>

`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker>
  <input type="text" class="am-form-field" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

### 参数设置

- format: 日期格式，默认为 `yyyy/mm/dd` ，符合中文日期选择，可以选择 `yy/mm/dd` 、`mm/dd` 或者英文日期格式 `dd/mm/yyyy`、`dd/mm/yy`、`dd/mm`等，中间分隔符可以使用 `/`、`-`、` `等。
- language: 语言选择，默认为中文 `zh`，目前支持两种：`zh`、`en` 语言选择，支持多语言，源码中扩展。

`````html
<p><input type="text" class="am-form-field am-radius" placeholder="默认日期格式" data-am-datepicker /></p>
<p><input type="text" class="am-form-field am-radius" placeholder="yy-mm-dd 日期格式" data-am-datepicker='{format: "yy-mm-dd"}' /></p>
<p><input type="text" class="am-form-field am-radius" placeholder="dd/mm/yyyy language:en" data-am-datepicker='{format: "dd/mm/yyyy", language: "en"}' /></p>

`````
```html
<p><input type="text" class="am-form-field am-radius" placeholder="默认日期格式" data-am-datepicker /></p>
<p><input type="text" class="am-form-field am-radius" placeholder="yy/mm/dd 日期格式" data-am-datepicker='{format: "yy-mm-dd"}' /></p>
<p><input type="text" class="am-form-field am-radius" placeholder="dd/mm/yyyy language:en" data-am-datepicker='{format: "dd/mm/yyyy", language: "en"}' /></p>
```
