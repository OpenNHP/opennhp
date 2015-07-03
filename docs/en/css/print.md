# Print
---

These are the printing styles. When a page is printed, `color` will be set to black and styles like `background`, `text-shadow` and `box-shadow` will be removed.

## Print URL of Links

Using `content` attribute in CSS3, title and URL of `<a>` and `<abbr>` will be printed after its content.

`````html
<a href="http://www.amazeui.org">Amaze UI</a>
<br/>
<abbr title="http://www.amazeui.org">Amaze UI</abbr>
`````

```html
<!-- Use print in browser to preview -->
<a href="http://www.amazeui.org">Amaze UI</a>
<abbr title="http://www.amazeui.org">Amaze UI</abbr>
```

## Utility Class

- `am-print-hide` Visible in browser, but Invisible to printer.
- `am-print-block`、`am-print-inline-block`、`am-print-inline` Invisible in browser, but visible to printer, and are printed as `block`,`inline-block`,`inline` styles respectively.


`````html
<div class="am-print-hide"><button type="button" class="am-btn am-btn-primary am-btn-block">Visible in browser, but Invisible to printer.</button></div>
<div class="am-print-block"><button type="button" class="am-btn am-btn-primary am-btn-block">Visible in browser, but Invisible to printer.</button></div>
`````

```html
<!-- 在打印预览中查看效果 -->
<div class="am-print-hide"><button type="button" class="am-btn am-btn-primary am-btn-block">Invisible in browser, but visible to printer.</button></div>

<div class="am-print-block"><button type="button" class="am-btn am-btn-primary am-btn-block">Invisible in browser, but visible to printer.</button></div>
```

## Reference

- [html5-boilerplate main.css](https://github.com/h5bp/html5-boilerplate/blob/master/src/css/main.css)
