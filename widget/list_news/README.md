# List 内容列表
---

内容列表组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `list_news`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
   // 列表标题
  "header": {
    "title": "最新文章",
    "link": "###",
    "moreText": "更多 >",
    "morePosition": "bottom"    // 更多链接位置 [bottom|top]
  },

  // 列表主要内容
  "main": [
    {
      "title": "",            // 新闻标题
      "link": "",             // 新闻链接
      "date": "",             // 日期
      "desc": "",             // 摘要信息，支持html
      "img": "",              // 缩略图地址

      // 数据接口
      "thumbAddition": "",    // 缩略图附加信息，可传 html，在thumb 模式下有效，显示在缩略图下面
      "mainAddition": ""      // 附加信息，可传 html，任何模式下都有效
    }
  ]
};

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
   // normal、thumb
    "type": "",

    // 当type为 thumb 时必传
    "thumbPosition": "bottom"
    // top-一般用于全宽模式
    // bottom - bottom-left | bottom-right
    // left - right
  },

  // 数据传递
  "content": {
    "header": {
      "title": "最新文章", // 栏目标题
      "link": "###",
      "className": "",
      "moreText": "更多 >", // 更多链接显示文字
      "morePosition": "bottom" // 【更多】链接位置，top、bottom，如果不设置则不显示更多链接
    },
    main: [
      {
        "title": "", // 新闻标题
        "link": "",  // 新闻链接
        "className": "",  // 自定 class
        "date": "",  // 日期
        "desc": "",  // 摘要信息，支持html
        "img": "", // 图片链接

        // 数据接口
        "thumbAddition": "", // 缩略图附加信息，可传 html，thumb 模式下有效，显示在缩略图下面
        "mainAddition": ""  // 附加信息，可传 html，任何模式下都有效
      }
    ]
}
```
