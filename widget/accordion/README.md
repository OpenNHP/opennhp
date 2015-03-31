# Accordion 折叠面板
---

手风琴折叠面板组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

首先，引入 Handlebars 库及 `amui.widget.helper.js`（可参考 [Amaze UI HTML 模板](/getting-started) 里的 `widget.html` 文件）。

接下来有两种方式来渲染数据。

__第一种，将组件调用代码写在模板里面：__

```html
<script type="text/x-handlebars-template" id="my-tpl">
  {{>accordion accordionData}}
</script>
```

然后获取模板内容，传入数据渲染并插入到页面中。

```javascript
$(function() {
  var $tpl = $('#my-tpl'),
      tpl = $tpl.text(),
      template = Handlebars.compile(tpl),
      data = {
        accordionData: {
          "theme": "basic",
          "content": [
            {
              "title": "标题一",
              "content": "内容一",
              "active": true
            },
            {
              "title": "标题二",
              "content": "内容二"
            },
            {
              "title": "标题三",
              "content": "内容三"
            }
          ]
        }
      },
      html = template(data);

  $tpl.before(html);
});
```

渲染出来的 HTML 如下：

```html
<section data-am-widget="accordion" class="am-accordion am-accordion-basic doc-accordion-class"
         id="doc-accordion-example" data-accordion-settings="{  }">
  <dl class="am-accordion-item am-active">
    <dt class="am-accordion-title">标题一</dt>
    <dd class="am-accordion-content">内容一</dd>
  </dl>
  <dl class="am-accordion-item">
    <dt class="am-accordion-title">标题二</dt>
    <dd class="am-accordion-content">内容二</dd>
  </dl>
  <dl class="am-accordion-item">
    <dt class="am-accordion-title">标题三</dt>
    <dd class="am-accordion-content">内容三</dd>
  </dl>
</section>
```

如果使用的组件较多或者还有组件以外的自定义模板，建议使用上面的方法，将模板分离出来，便于维护。

__第二种，直接将组件调用代码传给 Handlebars：__

```javascript
var template = Handlebars.compile('{{>accordion}}'),
    data = {
      accordionData: {
        "id": "doc-accordion-example",
        "className": "doc-accordion-class",
        "theme": "basic",
        "content": [
          {
            "title": "标题一",
            "content": "内容一",
            "active": true
          },
          {
            "title": "标题二",
            "content": "内容二"
          },
          {
            "title": "标题三",
            "content": "内容三"
          }
        ]
      }
    },
    html = template(data.accordionData);

$('body').append(html);
```

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = [
  {
    "title": "",    // 面板标题 支持 html
    "content": ""   // 面板内容 支持 html
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
    "multiple": false // 是否允许同时展开多个面板，默认为 FALSE
  },

  // 内容（*为必备项）
  "content": [
    {
      "title": "", // 标题，支持 html
      "content": "", // 内容，支持 html
      "active": false // 是否激活当前面板，如果需要激活则设置为 true，否则可不用设置此项
      // Amaze UI 2.3 新增
      "disabled": null // 是否禁用当前面板，如果需要禁用则设置为 true，否则不用设置此项，禁用以后此面板将保持默认状态，不响应用户操作
    }
  ]
}
```

## 注意事项

- **不要在 `.am-accordion-bd` 上添加上下 `padding`/`margin`/`border` 样式**。
