---
id: datepicker
title: 日期选择
titleEn: Datepicker
prev: javascript/tabs.html
next: javascript/selected.html
source: js/ui.datepicker.js
doc: docs/javascript/datepicker.md
---

# Datepicker
---

日期选择插件。**需要时间选择的参见**：[DateTimePicker - 日期时间选择插件](https://github.com/amazeui/datetimepicker)。

**注意：**

在触控设备上， `<input>` 获得焦点时会激活键盘，部分浏览器添加 `readonly` 属性以后可禁止键盘激活。

## 使用演示

### 基本形式

在 `<input>` 上增加 `data-am-datepicker` 属性，调用日期插件。

`````html
<form action="" class="am-form" data-am-validator>
  <p>
  <input type="text" class="am-form-field" placeholder="日历组件" data-am-datepicker readonly required />
  </p>
  <p><button class="am-btn am-btn-primary">提交</button></p>
</form>
`````
```html
<form action="" class="am-form" data-am-validator>
  <p>
  <input type="text" class="am-form-field" placeholder="日历组件" data-am-datepicker readonly required />
  </p>
  <p><button class="am-btn am-btn-primary">提交</button></p>
</form>
```

### 结合组件使用

结合 `.am-input-group` 使用，父类添加 class `.am-datepicker-date` ，非 `<input>` 触发元素需增加 `.am-datepicker-add-on` class。

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'dd-mm-yyyy'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'dd-mm-yyyy'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

### 更改颜色

默认为蓝色，设置 `theme` 选项可改变颜色：

- `success`: 绿色
- `warning`: 橙色
- `danger`: 红色

`````html
<p>
  <input type="text" class="am-form-field" placeholder="日历组件"
         data-am-datepicker="{theme: 'success', weekStart: 6}" readonly/>
</p>
`````
```html
<p><input type="text" class="am-form-field" placeholder="日历组件" data-am-datepicker="{theme: 'success'}" readonly/></p>
```

### 视图模式

通过参数 `viewMode` 设置日历初始视图模式：

- `days`: 显示天（默认）
- `months`: 显示月
- `years`: 显示年

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm-dd', viewMode: 'years'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>

`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm-dd', viewMode: 'years'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

### 限制视图模式

设置参数 `minViewMode` 可以限制视图模式。下面的示例中限制了只能选择到月份：

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm', viewMode: 'years', minViewMode: 'months'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>

`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm', viewMode: 'years', minViewMode: 'months'}">
  <input type="text" class="am-form-field" placeholder="日历组件" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

**只能选择年份：**

注意 `format: 'yyyy '` 里面 `yyyy ` 后面多加一个空格。

`````html
<div>
  <input type="text" class="am-form-field" data-am-datepicker="{format: 'yyyy ', viewMode: 'years', minViewMode: 'years'}" placeholder="日历组件" data-am-datepicker readonly/>
</div>
`````
```html
<div>
  <input type="text" class="am-form-field" data-am-datepicker="{format: 'yyyy ', viewMode: 'years', minViewMode: 'years'}" placeholder="日历组件" data-am-datepicker readonly/>
</div>
```


### 自定义事件

通过监听自定义事件 `changeDate`，可以在回调函数中进行验证等操作。通过 `$().data('date')` 获取改变后的日期。

`````html
<div class="am-alert am-alert-danger" id="my-alert" style="display: none">
  <p>开始日期应小于结束日期！</p>
</div>
<div class="am-g">
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-start">开始日期</button><span id="my-startDate">2014-12-20</span>
  </div>
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-end">结束日期</button><span id="my-endDate">2014-12-25</span>
  </div>
</div>
<script>
  $(function() {
    var startDate = new Date(2014, 11, 20);
    var endDate = new Date(2014, 11, 25);
    var $alert = $('#my-alert');
    $('#my-start').datepicker().
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() > endDate.valueOf()) {
          $alert.find('p').text('开始日期应小于结束日期！').end().show();
        } else {
          $alert.hide();
          startDate = new Date(event.date);
          $('#my-startDate').text($('#my-start').data('date'));
        }
        $(this).datepicker('close');
      });

    $('#my-end').datepicker().
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() < startDate.valueOf()) {
          $alert.find('p').text('结束日期应大于开始日期！').end().show();
        } else {
          $alert.hide();
          endDate = new Date(event.date);
          $('#my-endDate').text($('#my-end').data('date'));
        }
        $(this).datepicker('close');
      });
  });
</script>
`````

```html
<div class="am-alert am-alert-danger" id="my-alert" style="display: none">
  <p>开始日期应小于结束日期！</p>
</div>
<div class="am-g">
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-start">开始日期</button><span id="my-startDate">2014-12-20</span>
  </div>
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-end">结束日期</button><span id="my-endDate">2014-12-25</span>
  </div>
</div>
<script>
  $(function() {
    var startDate = new Date(2014, 11, 20);
    var endDate = new Date(2014, 11, 25);
    var $alert = $('#my-alert');
    $('#my-start').datepicker().
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() > endDate.valueOf()) {
          $alert.find('p').text('开始日期应小于结束日期！').end().show();
        } else {
          $alert.hide();
          startDate = new Date(event.date);
          $('#my-startDate').text($('#my-start').data('date'));
        }
        $(this).datepicker('close');
      });

    $('#my-end').datepicker().
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() < startDate.valueOf()) {
          $alert.find('p').text('结束日期应大于开始日期！').end().show();
        } else {
          $alert.hide();
          endDate = new Date(event.date);
          $('#my-endDate').text($('#my-end').data('date'));
        }
        $(this).datepicker('close');
      });
  });
</script>
```

### 设置禁止选择日期

初始化的时候通过 `onRender` 选项设置禁用日期。

**`v2.5`**：`onRender` 方法增加了 `viewMode` 参数，以便更准确的处理不同视图渲染。

`viewMode` 内部调用时使用了以下值：

- `0`: 天视图
- `1`: 月视图
- `2`: 年视图

`````html
<div class="am-g">
  <div class="am-u-sm-6">
    设置禁用日期<br/>
    <p><input type="text" class="am-form-field" placeholder="今天之前的日期被禁用" id="my-start-2"/></p>
  </div>
  <div class="am-u-sm-6">
    禁用日期<br/>
    <p><input type="text" class="am-form-field" id="my-end-2" /></p>
  </div>
</div>
<script>
  $(function() {
    var nowTemp = new Date();
    var nowDay = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), nowTemp.getDate(), 0, 0, 0, 0).valueOf();
    var nowMoth = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), 1, 0, 0, 0, 0).valueOf();
    var nowYear = new Date(nowTemp.getFullYear(), 0, 1, 0, 0, 0, 0).valueOf();
    var $myStart2 = $('#my-start-2');

    var checkin = $myStart2.datepicker({
      onRender: function(date, viewMode) {
        // 默认 days 视图，与当前日期比较
        var viewDate = nowDay;

        switch (viewMode) {
          // moths 视图，与当前月份比较
          case 1:
            viewDate = nowMoth;
            break;
          // years 视图，与当前年份比较
          case 2:
            viewDate = nowYear;
            break;
        }

        return date.valueOf() < viewDate ? 'am-disabled' : '';
      }
    }).on('changeDate.datepicker.amui', function(ev) {
        if (ev.date.valueOf() > checkout.date.valueOf()) {
          var newDate = new Date(ev.date)
          newDate.setDate(newDate.getDate() + 1);
          checkout.setValue(newDate);
        }
        checkin.close();
        $('#my-end-2')[0].focus();
    }).data('amui.datepicker');

    var checkout = $('#my-end-2').datepicker({
      onRender: function(date, viewMode) {
        var inTime = checkin.date;
        var inDay = inTime.valueOf();
        var inMoth = new Date(inTime.getFullYear(), inTime.getMonth(), 1, 0, 0, 0, 0).valueOf();
        var inYear = new Date(inTime.getFullYear(), 0, 1, 0, 0, 0, 0).valueOf();

        // 默认 days 视图，与当前日期比较
        var viewDate = inDay;

        switch (viewMode) {
          // moths 视图，与当前月份比较
          case 1:
            viewDate = inMoth;
            break;
          // years 视图，与当前年份比较
          case 2:
            viewDate = inYear;
            break;
        }

        return date.valueOf() <= viewDate ? 'am-disabled' : '';
      }
    }).on('changeDate.datepicker.amui', function(ev) {
      checkout.close();
    }).data('amui.datepicker');
  });
</script>
`````
```html
<div class="am-g">
  <div class="am-u-sm-6">
    设置禁用日期<br/>
    <p><input type="text" class="am-form-field" placeholder="今天之前的日期被禁用" id="my-start-2"/></p>
  </div>
  <div class="am-u-sm-6">
    禁用日期<br/>
    <p><input type="text" class="am-form-field" id="my-end-2" /></p>
  </div>
</div>
<script>
  $(function() {
    var nowTemp = new Date();
    var nowDay = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), nowTemp.getDate(), 0, 0, 0, 0).valueOf();
    var nowMoth = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), 1, 0, 0, 0, 0).valueOf();
    var nowYear = new Date(nowTemp.getFullYear(), 0, 1, 0, 0, 0, 0).valueOf();
    var $myStart2 = $('#my-start-2');

    var checkin = $myStart2.datepicker({
      onRender: function(date, viewMode) {
        // 默认 days 视图，与当前日期比较
        var viewDate = nowDay;

        switch (viewMode) {
          // moths 视图，与当前月份比较
          case 1:
            viewDate = nowMoth;
            break;
          // years 视图，与当前年份比较
          case 2:
            viewDate = nowYear;
            break;
        }

        return date.valueOf() < viewDate ? 'am-disabled' : '';
      }
    }).on('changeDate.datepicker.amui', function(ev) {
        if (ev.date.valueOf() > checkout.date.valueOf()) {
          var newDate = new Date(ev.date)
          newDate.setDate(newDate.getDate() + 1);
          checkout.setValue(newDate);
        }
        checkin.close();
        $('#my-end-2')[0].focus();
    }).data('amui.datepicker');

    var checkout = $('#my-end-2').datepicker({
      onRender: function(date, viewMode) {
        var inTime = checkin.date;
        var inDay = inTime.valueOf();
        var inMoth = new Date(inTime.getFullYear(), inTime.getMonth(), 1, 0, 0, 0, 0).valueOf();
        var inYear = new Date(inTime.getFullYear(), 0, 1, 0, 0, 0, 0).valueOf();

        // 默认 days 视图，与当前日期比较
        var viewDate = inDay;

        switch (viewMode) {
          // moths 视图，与当前月份比较
          case 1:
            viewDate = inMoth;
            break;
          // years 视图，与当前年份比较
          case 2:
            viewDate = inYear;
            break;
        }

        return date.valueOf() <= viewDate ? 'am-disabled' : '';
      }
    }).on('changeDate.datepicker.amui', function(ev) {
      checkout.close();
    }).data('amui.datepicker');
  });
</script>
```

## 调用方式

### 通过 Data API

添加 `data-am-datepicker` 属性，并设置相关选项。

```html
<input class="" data-am-datepicker="{format: 'yyyy-mm'}"/>
```

#### JS 调用

通过 `$().datepicker(options)` 调用。

```javascript
$('#my-datepicker').datepicker({format: 'yyyy-mm'});
```

#### 方法说明

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>方法名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.datepicker('open')</code></td>
    <td>显示日历</td>
  </tr>
  <tr>
    <td><code>.datepicker('close')</code></td>
    <td>隐藏日历</td>
  </tr>
  <tr>
    <td><code>.datepicker('place')</code></td>
    <td>更新调用<code>datepicker</code>的相对位置</td>
  </tr>
  <tr>
    <td><code>.datepicker('setValue', value)</code></td>
    <td>设置<code>Datepicker</code>新值</td>
  </tr>
  </tbody>
</table>

#### 选项说明

- `format`: 日期格式，默认为 `yyyy-mm-dd`，可以选择 `yy/mm/dd` 、`mm/dd`、`dd/mm/yyyy`、`dd/mm/yy`、`dd/mm`等，中间分隔符可以使用 `/`、`-`、` `。
- `viewMode`: 日期选择器初始视图模式，`string`|`integer`， 默认为 0，可选值 `days`、`months`、`years`或者对应的 `0`、`1`、`2`。
- `minViewMode`: 日期选择器初始视图模式限制，`string`|`integer`， 默认为 `0`，可选值`days`、`months`、`years`或者对应的 `0`、`1`、`2`。
- `onRender`: 渲染日历时调用的函数，比如 `.am-disabled` 设置禁用日期。
- `theme`: 设置日期颜色主题，可选值为 `success`、`danger`、`warning`，对应为绿色、红色、橙色，默认为蓝色。
- `locale`: 语言设置, 可选值为 `zh_CN`、`en_US`，默认为中文。
- `autoClose`: 日期选定以后是否自动关闭日期选择器, 默认为 `true` (仅在 `days` 视图有效)。

设置 `viewMode` 和 `minViewMode` 需要注意日期格式 `format` 的设置。

#### 事件说明

选择日期时，通过查看控制台选择的日期。

`````html
<p><input type="text" class="am-form-field" placeholder="日历组件" id="doc-datepicker"/></p>
<script>
$(function() {
  $('#doc-datepicker').datepicker().
    on('changeDate.datepicker.amui', function(event) {
      console.log(event.date);
    });
});
</script>
`````

```javascript
$(function() {
  $('#doc-datepicker').datepicker().
    on('changeDate.datepicker.amui', function(event) {
      console.log(event.date);
    });
});
```

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>changeDate.datepicker.amui</code></td>
    <td>日期改变时触发</td>
  </tr>
  </tbody>
</table>

### 语言扩展

内置英语和简体中文支持，默认为中文，要支持更多语言可以通过 `Datepicker.locales` 扩展。

**设置语言：**

`````html
<p>
  <input type="text" class="am-form-field" placeholder="YYYY-MM-DD"
         data-am-datepicker="{locale: 'en_US'}" readonly/>
</p>
`````
```html
<p>
  <input type="text" class="am-form-field" placeholder="YYYY-MM-DD"
         data-am-datepicker="{locale: 'en_US'}" readonly/>
</p>
```

**扩展语言：**

<script>
(function($) {
  $.AMUI && $.AMUI.datepicker && ($.AMUI.datepicker.locales.fr = {
    days: ["Dimanche", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi"],
    daysShort: ["Dim", "Lun", "Mar", "Mer", "Jeu", "Ven", "Sam"],
    daysMin: ["D", "L", "Ma", "Me", "J", "V", "S"],
    months: ["Janvier", "Février", "Mars", "Avril", "Mai", "Juin", "Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre"],
    monthsShort: ["Jan", "Fev", "Mar", "Avr", "Mai", "Jui", "Jul", "Aou", "Sep", "Oct", "Nov", "Dec"],
    weekStart: 1
  });
})(window.jQuery);
</script>

`````html
<p>
  <input type="text" class="am-form-field" placeholder="来一丢丢 French"
         data-am-datepicker="{locale: 'fr', autoClose: 0}" readonly/>
</p>
`````
```html
<p>
  <input type="text" class="am-form-field" placeholder="来一丢丢 French"
         data-am-datepicker="{locale: 'fr', autoClose: 0}" readonly/>
</p>

<script>
(function($) {
  $.AMUI && $.AMUI.datepicker && ($.AMUI.datepicker.locales.fr = {
    days: ["Dimanche", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"],
    daysShort: ["Dim", "Lun", "Mar", "Mer", "Jeu", "Ven", "Sam", "Dim"],
    daysMin: ["D", "L", "Ma", "Me", "J", "V", "S", "D"],
    months: ["Janvier", "Février", "Mars", "Avril", "Mai", "Juin", "Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre"],
    monthsShort: ["Jan", "Fev", "Mar", "Avr", "Mai", "Jui", "Jul", "Aou", "Sep", "Oct", "Nov", "Dec"],
    weekStart: 1
  });
})(window.jQuery);
</script>
```
