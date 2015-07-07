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

A plugin used to select date. If you need timepicker, see [DateTimePicker](https://github.com/amazeui/datetimepicker).

**Attention: **

On touch screen devices, `<input>` will activate keyboard when it is focused. Keyboard activation can be disabled by adding `readonly` in some browsers. 

## Examples

### Basic

To create a datepicker, add `.data-am-datepicker` attribute to `<input>` element.

`````html
<p><input type="text" class="am-form-field" placeholder="日历组件" data-am-datepicker="{locale: 'en_US'}" readonly/></p>

`````
```html
<p><input type="text" class="am-form-field" placeholder="日历组件" data-am-datepicker="{locale: 'en_US'}" readonly/></p>
```

### With Other Components

Datepicker can be used with `.am-input-group`. Add the `.am-datepicker-date` class to parent element, or add the `.am-datepicker-add-on` class if not using `<input>` element as trigger element.

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'dd-mm-yyyy', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'dd-mm-yyyy', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

### Color

Default color is blue. Set `theme` option to change color:

- `success`: Green
- `warning`: Orange
- `danger`: Red

`````html
<p>
  <input type="text" class="am-form-field" placeholder="Calendar"
         data-am-datepicker="{theme: 'success', weekStart: 6, locale: 'en_US'}" readonly/>
</p>
`````
```html
<p><input type="text" class="am-form-field" placeholder="Calendar" data-am-datepicker="{theme: 'success', locale: 'en_US'}" readonly/></p>
```

### View Mode

Use `viewMode` option to change the initial view mode:

- `days`: Show days (Default)
- `months`: Show months
- `years`: Show years

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm-dd', viewMode: 'years', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>

`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm-dd', viewMode: 'years', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

### Minimum View Mode

Use `minViewMode` to set the minimum view mode. In the following example, the minimum view mode is set to months.

`````html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm', viewMode: 'years', minViewMode: 'months', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>

`````
```html
<div class="am-input-group am-datepicker-date" data-am-datepicker="{format: 'yyyy-mm', viewMode: 'years', minViewMode: 'months', locale: 'en_US'}">
  <input type="text" class="am-form-field" placeholder="Calendar" readonly>
  <span class="am-input-group-btn am-datepicker-add-on">
    <button class="am-btn am-btn-default" type="button"><span class="am-icon-calendar"></span> </button>
  </span>
</div>
```

**Years Only:**

Please be aware of that there is one more space at the end of `'yyyy '` in ` `format: 'yyyy '`.

`````html
<div>
  <input type="text" class="am-form-field" data-am-datepicker="{format: 'yyyy ', viewMode: 'years', minViewMode: 'years', locale: 'en_US'}" placeholder="Calendar" data-am-datepicker readonly/>
</div>
`````
```html
<div>
  <input type="text" class="am-form-field" data-am-datepicker="{format: 'yyyy ', viewMode: 'years', minViewMode: 'years', locale: 'en_US'}" placeholder="Calendar" data-am-datepicker readonly/>
</div>
```


### Events

Varification can be done in the callback function of event `changeDate`. Use `$().data('date')` to get the date.

`````html
<div class="am-alert am-alert-danger" id="my-alert" style="display: none">
  <p>Start date should be earlier than end date!</p>
</div>
<div class="am-g">
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-start">Start in</button><span id="my-startDate">2014-12-20</span>
  </div>
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-end">End in</button><span id="my-endDate">2014-12-25</span>
  </div>
</div>
<script>
  $(function() {
    var startDate = new Date(2014, 11, 20);
    var endDate = new Date(2014, 11, 25);
    var $alert = $('#my-alert');
    $('#my-start').datepicker({locale: 'en_US'}).
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() > endDate.valueOf()) {
          $alert.find('p').text('Start date should be earlier than end date!').end().show();
        } else {
          $alert.hide();
          startDate = new Date(event.date);
          $('#my-startDate').text($('#my-start').data('date'));
        }
        $(this).datepicker('close');
      });

    $('#my-end').datepicker({locale: 'en_US'}).
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() < startDate.valueOf()) {
          $alert.find('p').text('Start date should be earlier than end date!').end().show();
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
  <p>Start date should be earlier than end date!</p>
</div>
<div class="am-g">
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-start">Start in</button><span id="my-startDate">2014-12-20</span>
  </div>
  <div class="am-u-sm-6">
    <button type="button" class="am-btn am-btn-default am-margin-right" id="my-end">End in</button><span id="my-endDate">2014-12-25</span>
  </div>
</div>
<script>
  $(function() {
    var startDate = new Date(2014, 11, 20);
    var endDate = new Date(2014, 11, 25);
    var $alert = $('#my-alert');
    $('#my-start').datepicker({locale: 'en_US'}).
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() > endDate.valueOf()) {
          $alert.find('p').text('Start date should be earlier than end date!').end().show();
        } else {
          $alert.hide();
          startDate = new Date(event.date);
          $('#my-startDate').text($('#my-start').data('date'));
        }
        $(this).datepicker('close');
      });

    $('#my-end').datepicker({locale: 'en_US'}).
      on('changeDate.datepicker.amui', function(event) {
        if (event.date.valueOf() < startDate.valueOf()) {
          $alert.find('p').text('Start date should be earlier than end date!').end().show();
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

### Avaliable dates

Set the avaliable dates with `onRender` option.

`````html
<div class="am-g">
  <div class="am-u-sm-6">
    Set avaliable dates<br/>
    <p><input type="text" class="am-form-field" placeholder="Only dates after today is avaliable" id="my-start-2"/></p>
  </div>
  <div class="am-u-sm-6">
    Avaliable dates<br/>
    <p><input type="text" class="am-form-field" id="my-end-2" /></p>
  </div>
</div>
<script>
  $(function() {
    var nowTemp = new Date();
    var now = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), nowTemp.getDate(), 0, 0, 0, 0);
    var $myStart2 = $('#my-start-2');

    var checkin = $myStart2.datepicker({
      onRender: function(date) {
        return date.valueOf() < now.valueOf() ? 'am-disabled' : '';
      },
      locale: 'en_US'
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
      onRender: function(date) {
        return date.valueOf() <= checkin.date.valueOf() ? 'am-disabled' : '';
      },
      locale: 'en_US'
    }).on('changeDate.datepicker.amui', function(ev) {
      checkout.close();
    }).data('amui.datepicker');

  })
</script>
`````

```html
<div class="am-g">
  <div class="am-u-sm-6">
    Set avaliable dates<br/>
    <p><input type="text" class="am-form-field" placeholder="Only dates after today is avaliable" id="my-start-2"/></p>
  </div>
  <div class="am-u-sm-6">
    Avaliable dates<br/>
    <p><input type="text" class="am-form-field" id="my-end-2" /></p>
  </div>
</div>
<script>
  $(function() {
    var nowTemp = new Date();
    var now = new Date(nowTemp.getFullYear(), nowTemp.getMonth(), nowTemp.getDate(), 0, 0, 0, 0);
    var $myStart2 = $('#my-start-2');

    var checkin = $myStart2.datepicker({
      onRender: function(date) {
        return date.valueOf() < now.valueOf() ? 'am-disabled' : '';
      },
      locale: 'en_US'
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
      onRender: function(date) {
        return date.valueOf() <= checkin.date.valueOf() ? 'am-disabled' : '';
      },
      locale: 'en_US'
    }).on('changeDate.datepicker.amui', function(ev) {
      checkout.close();
    }).data('amui.datepicker');

  })
</script>
```

## Usage

### Using Data API

Add `data-am-datepicker` attribute, and set options in it.

```html
<input class="" data-am-datepicker="{format: 'yyyy-mm'}"/>
```

#### Using JS

Use Method `$().datepicker(options)`.

```javascript
$('#my-datepicker').datepicker({format: 'yyyy-mm'});
```

#### Instruction

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Method</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.datepicker('open')</code></td>
    <td>Show calendar</td>
  </tr>
  <tr>
    <td><code>.datepicker('close')</code></td>
    <td>Hide calendar</td>
  </tr>
  <tr>
    <td><code>.datepicker('place')</code></td>
    <td>Refresh the relative postion of <code>datepicker</code></td>
  </tr>
  <tr>
    <td><code>.datepicker('setValue', value)</code></td>
    <td>Set new value to <code>Datepicker</code></td>
  </tr>
  </tbody>
</table>

#### Options

- `format`: Date format. Default format is `yyyy-mm-dd`. Other formats include `yy/mm/dd`, `mm/dd`, `dd/mm/yyyy`, `dd/mm/yy`, `dd/mm` and etc. Dividers are chosen from `/`, `-` and ` `.
- `viewMode`: Initial view mode. `string`|`integer`. Default value is 0. Avaliable options include `'days'`, `'months'` and `'years'`, or respective `0`, `1` and `2`.
- `minViewMode`: Minimum view mode. `string`|`integer`. Default value is `0`.Options include `'days'`, `'months'` and `'years'`, or respective `0`, `1` and `2`.
- `onRender`: The function get called when rendering. For example, use `.am-disabled` to set the disabled and avaliable dates.
- `theme`: Theme of datepicker. Options include `success`, `danger` and `warning`, which use green, red and orange respectively. The default color is blue.
- `locale`: Language. Options include `zh_CN` and `en_US`. Default language is Chinese.
- `autoClose`: Close the datepicker when date is selected. Default value is `true` (Only works in `days` view).

设置 `viewMode` 和 `minViewMode` 需要注意日期格式 `format` 的设置。

#### Events

Print log in console when selecting date.

`````html
<p><input type="text" class="am-form-field" placeholder="Calendar" id="doc-datepicker"/></p>
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
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>changeDate.datepicker.amui</code></td>
    <td>Triggered when selected date is changed</td>
  </tr>
  </tbody>
</table>

### Language Extension

Support English and Chinese (Simplified). Default language is Chinese. Use `Datepicker.locales` to add support to more languages.

**Language Setting：**

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

**Language Extension：**

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
