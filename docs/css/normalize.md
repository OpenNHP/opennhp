# Normalize
---

在统一浏览器默认样式上，[Reset](http://meyerweb.com/eric/tools/css/reset/ ) 一度非常流行，更有简单粗暴的通配符 reset ：

```css
* {
    margin: 0;
    padding: 0;
    border:0;
}
```

时过境迁，Reset 逐渐淡出的前沿前端的视野，[normalize.css](https://github.com/necolas/normalize.css) 取而代之。normalize.css，统一样式的同时保留可辨识性；reset 统一样式，完全没有可读性，分不清是男人、女人，或者是不男不女，看着都一样。

Amaze UI 也使用了 normalize.css，但部分细节做了一些调整：

- `html` 添加 `-webkit-font-smoothing: antialiased;`
- `<hgroup>` 已经从 W3C 标准中移除，不建议使用
- `<img>` 设置最大宽度为 `100%`
- `<figure>` 外边距设置为 `0`
- `<textarea>` 添加 `vertical-align: top; resize: vertical;`
- 移除 `<dfn>` 斜体字样式
- 移除 `<h1>` 样式