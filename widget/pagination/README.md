# Pagination 分页
---

___本组件样式有待改善!___

分页组件。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `figure`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
  "prevTitle": "上一页",  //（可选）内容可以填写成其它的内容
  "prevLink": "#",        //（可选）直接填写原网站中a链接的href地址

  "nextTitle": "下一页",  //（可选）内容可以填写成其它的内容
  "nextLink": "#",        //（可选）直接填写原网站中a链接的href地址

  "firstTitle": "第一页", //（可选）内容可以填写成其它的内容
  "firstLink": "#",       //（可选）直接填写原网站中a链接的href地址

  "lastTitle": "最末页",  //（可选）内容可以填写成其它的内容
  "lastLink": "#",        //（可选）直接填写原网站中a链接的href地址

  "total": "",            // （可选，赋值是“3/5”样式，否则是“3”样式）显示总的页数

  "page": [
    {
      "title": "1",
      "link": "#"     //直接填写原网站中a链接的href地址，这里也可以用一个循环将原页面中的123456...页的链接地址添加进来。
    },
    {
      "title": "2",
      "link": "#"
    }
  ]
};

return data;
```

### 设置说明（0.9）

- `options` 内 `select` 选项：默认是 theme-1，更改此属性后，使用 theme-2 或 theme-3；
- 设置总页数 `total` 值：显示形式为 `3/5`，否则为 `3` 形式；
- 使用 `theme-2` 或 `theme-3` 时，页码改变调用函数为：pageChange()，需自行定义。

## 数据结构

```javascript
{
  "id": "",

  // 自定义class
  "className": "",

  // 主题
  "theme": "default",

  "options": {
    "select": ""
  },

  "content": {
    // 上一页
    "prevTitle": "上一页",
    "prevLink": "#",

    // 第一页
    "firstTitle": "第一页",
    "firstLink": "#",

    // 可选
    "nextTitle": "下一页",
    "nextLink": "#",

    // 可选
    "lastTitle": "最末页",
    "lastLink": "#",

    "total": "", // 总页数

    "page": [
      {
        "title": "1",
        "link": "#",
        "class": "" // AMUI 0.9
        "className": "", // AMUI 1.0
      }
    ]
  }
}
```
