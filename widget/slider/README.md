# Slider
---

图片轮播模块，更多细节[参见 JS 插件中的介绍](/javascript/slider?_ver=2.x)。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `slider`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var content = [
  {
    "img": "",      // 表示轮播图片的路径，如：xxx.src() ，必传
    "link": "",     // 链接，可选
    "thumb": "",    // 如果需要，添加缩略图，填写缩略图的地址
    "desc": ""      // 当前图片的附加信息，支持 HTML，为高级定制提供 HTML 接口
  }
];

return content;
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

  // 配置，根据需求进行设置，传递 JSON 字符串
  "sliderConfig": "{}",

  //内容（*为必备项）
  "content": [
    {
      "img": "",
      "link": "", // 链接
      "thumb": "", // 缩略图
      "desc": "" // 附加信息，支持DOM，为高级定制提供DOM接口
    }
  ]
}
```

## Slider 参数说明

注意：通过模块传递的JSON字符串参数无法传递 callback，如需传递 callback 函数，请手动启动 slider：

- 添加 `am-slider-manual` 类名，禁用模块默认的初始化函数；
- 在自定义脚本里调用 slider：

```javascript
$(function() {
  $('.am-slider-manual').flexslider({
  // options
  });
});
```
