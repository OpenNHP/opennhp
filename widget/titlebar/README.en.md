# Titlebar 标题栏
---

标题栏组件，常用作页头、标题等。

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `titlebar`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "title": "",		// 主标题
  "link": "",       // 主标题链接，可选
  "nav": [          // 右侧附加链接，可以为多个
    {
      "link": "",
      "title": "",
      "className": ""
    }
  ]
};

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": null,

  "content": {
    "title": "",
    "link": "",
    "className": "",
    "nav": [
      {
        "link": "",
        "title": "",
        "className": ""
      }
    ]
  }
}
```
