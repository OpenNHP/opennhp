# Pagination 分页
---

___本组件样式有待改善!___

分页组件。

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `pagination`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

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
        "className": ""
      }
    ]
  }
}
```
