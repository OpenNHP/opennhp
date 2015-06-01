# Paragraph 段落
---

文本段落组件，可用于放置文本、图片、链接等。

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `paragraph`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

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
