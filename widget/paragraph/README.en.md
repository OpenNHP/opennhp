# Paragraph 段落
---

文本段落组件，可用于放置文本、图片、链接等。

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `paragraph`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "content": ""   // 填写 paragraph 的内容
};

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
    "imgLightbox": false, // 图片查看器
    "tableScrollable": false // 表格横向滚动功能
  },

  "content": {
    "content": ""
  }
}
```
