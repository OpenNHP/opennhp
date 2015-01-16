# Intro 简介组件
---

简介组件，常用于放置包含图片和文字的企业介绍信息等。


## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `intro`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
  "title": "",  // 标题
  "more": { // 更多链接
    "title": "",
    "link": ""
  },
  "left": "", //左边的内容
  "right": "" //右边的内容
};
return data;
```

## 数据接口

```javascript
{
	"id": "",

  "className": "",

	"theme": "",

	"options": {
		"leftCols": "",
		"RightCols": "",
		"position":""
	},

	"content": {
		"title": "",
		"more": {
      "link": "",
      "title": "",
      "className": ""
    },
		"left": "",
		"right": ""
	}
}
```
