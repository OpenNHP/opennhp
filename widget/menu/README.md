# Menu 菜单
---

菜单组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `menu`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

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
