# Menu 菜单
---

菜单组件。

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `menu`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title": "",            // 一级菜单标题
    "link": "",             // 一级菜单链接
    "className": "",
    "subMenu": [
      {                   // 二级菜单
        "title": "",    // 二级菜单标题
        "link": "",     // 二级菜单链接
        "target": ""
      }
    ],
    "subCols": 3            // 设置二级菜单列数
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
    "cols": 5, // 一级菜单列数 [1-12]

    "offCanvasFlip": false, // 仅在 offcanvas 主题下有效，侧滑菜单默认在左侧，如果要在右侧显示，请传递 true
  },

  // 菜单数据
  "content": [
    {
      "title": "", // 一级菜单标题
      "link": "", // 一级菜单链接
      "className": "", // 菜单 li 上的自定义 Class
      "subCols": "", // 子菜单(第二级)列数
      "channelLink": "进入栏目 &raquo;", // 如果设置了，二级菜单里将显示此文本并链接到该栏目
      "subMenu": [{
        "title": "", // 二级菜单标题
        "link": "", // 二级菜单链接
        "className": "", // 菜单上的 Class
      }]
    }
  ]
}
```
