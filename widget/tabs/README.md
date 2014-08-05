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
    "header" : "",      //每一个选项的标题
    "substance" : "",   //内容部分
    "class" : ""        //显示当前默认被选中的选项卡标题以及其内容，值是"active"，如果有多个选项，只允许一个tab被激活
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
    "cols": "" // 列数
  },

  // 每一个 Tab 的内容
	"content": [
    {
      "header": "",
      "substance": "",
      "class": ""
    }
  ]
}
```