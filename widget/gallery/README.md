# Gallery 图片画廊
---

图片画廊组件，用于展示图片为主体的内容。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `gallery`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = [
  {
    "img": "",      // 图片地址
    "link": "",     // 填写点击图片弹出放大时的图片路径
    "title": "",    // 图片标题
    "desc": ""      // 附加信息，支持DOM，为高级定制提供DOM接口
  }
];
return data;
```

## 数据接口

```javascript
{
  // id
  "id": "",

  // 自定义 class
  "className": "",

  // 主题
  "theme": "",

  "options": {
    "cols": 1,  // 列数
    "gallery": false // 是否开启点击图片全屏显示大图功能
  },

  //内容（*为必备项）
  "content": [
    {
      "img": "",
      "link": "", // 链接
      "title": "", // 图片标题
      "desc": "" // 附加信息，支持DOM，为高级定制提供DOM接口
    }
  ]
}
```
