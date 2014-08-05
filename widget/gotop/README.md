# GoTop 回到顶部
---

回到顶部组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `gotop`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
  "img": "",		// 不建议使用，可设置成背景
  "hint": "回到顶部"	// 显示文字
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
    "layout": "fixed"
  },

  "content": {
    "img": "",     // 一般情况下，不直接传图片，建议把图标放背景里
    "hint": "回到顶部"
  }
}
```