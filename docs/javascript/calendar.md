# Calendar
---

简介：这是一个“日历”插件，目前实现了一些基础功能。

### 基本使用

##### HTML代码
```html
<input type="text" id="calendar" value="" />
```

##### JAVASCRIPT代码
```javascript
$("#calendar").appendDtpicker();
```

### 参数设置
```html
/**
 * 可设置以下四个参数：
 * theme: 可写：iOS7，也可自定义（am-calendar-*）
 * lang: 默认中文，可写 en
 * position: 默认紧邻输入框，可设置为 bottom
 * futureOnly: 默认都显示，可设置为 true，只对未来时间有效
 */

/**
 * 使用示例一：设置主题为“iOS7”
 */
$("#calendar").appendDtpicker({"theme":"iOS7"});

/**
 * 使用示例二：设置语言为“英文”
 */
$("#calendar").appendDtpicker({"lang":"en"});

/**
 * 使用示例三：设置显示位置为“屏幕底部”
 */
$("#calendar").appendDtpicker({"position":"bottom"});

/**
 * 使用示例四：设置只对“未来时间”有效
 */
$("#calendar").appendDtpicker({"futureOnly":true});

/**
 * 使用示例五：同时设置多个参数
 */
$("#calendar").appendDtpicker({"lang":"en","futureOnly":true});
```