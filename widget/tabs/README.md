# Tabs 选项卡
---

选项卡组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `tabs`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

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
