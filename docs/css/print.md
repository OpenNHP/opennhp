# Print
---

打印样式组件，打印时将 `color` 设置成黑色，删除 `background` 、`text-shadow` 、`box-shadow` 样式，以节省打印机耗材，加快打印速度。

## 显示 URL 链接

利用 CSS3 `content` 属性，将 `<a>` 和 `<abbr>` 的标题、链接显示在后面。

`````html
<a href="http://www.amazeui.org">Amaze UI</a>
<br/>
<abbr title="http://www.amazeui.org">Amaze UI</abbr>
`````

```html
<!-- 点击浏览器菜单里的【打印】，预览打印样式 -->
<a href="http://www.amazeui.org">Amaze UI</a>
<abbr title="http://www.amazeui.org">Amaze UI</abbr>
```

## 辅助 Class

- `am-print-hide` 浏览器可见，打印机隐藏。
- `am-print-block`、`am-print-inline-block`、`am-print-inline` 打印机可见，浏览器隐藏。在打印机中分别对应为`block`,`inline-block`,`inline`样式。


`````html
<div class="am-print-hide"><button type="button" class="am-btn am-btn-primary am-btn-block">浏览器可见，打印机不可见</button></div>
<div class="am-print-block"><button type="button" class="am-btn am-btn-primary am-btn-block">打印机可见，浏览器不可见</button></div>
`````

```html
<!-- 在打印预览中查看效果 -->
<div class="am-print-hide"><button type="button" class="am-btn am-btn-primary am-btn-block">浏览器可见，打印机不可见</button></div>

<div class="am-print-block"><button type="button" class="am-btn am-btn-primary am-btn-block">打印机可见，浏览器不可见</button></div>
```

## 参考链接

- [html5-boilerplate main.css](https://github.com/h5bp/html5-boilerplate/blob/master/src/css/main.css)
