# Tabs 选项卡
---

选项卡组件。

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `tabs`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title":   "",   // 选项卡标题
    "content": "",   // 选项卡内容
    "active":  false // 是否激活当前选项卡，true | false，只允许一个 Tab 标记为激活
  }
];

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": {
    "noSwipe": false; // 是否禁用触控操作
  }

  "content": [
    {
      "title": "",
      "content": "",
      "active": ""
    }
  ]
}
```

## 常见问题

### 如何禁用触控操作？

在容器上添加 `data-am-tabs-noswipe="1"`。

```html
<div data-am-widget="tabs" class="am-tabs am-tabs-default" data-am-tabs-noswipe="1">
```
