# Code

Defines styles for display codes.

## Inline Code

Use `<code>` tag.

```html
<code>code here</code>
```

## Code Snippet

Put snippet in `<pre>`.

`````html
<pre style="background-color: #f8f8f8; color: #555">
window.addEventListener('load', function() {
    FastClick.attach(document.body);
}, false);
</pre>
`````

```html
<pre>
window.addEventListener('load', function() {
    FastClick.attach(document.body);
}, false);
</pre>
```

## 代码块高度

添加 `.am-pre-scrollable` 限制代码块高度，默认为 `24rem`。

`````html
<pre class="am-pre-scrollable" style="background-color: #f8f8f8; color: #555">
span.l-1 {-webkit-animation-delay: 1s;animation-delay: 1s;-ms-animation-delay: 1s;-moz-animation-delay: 1s;}
span.l-2 {-webkit-animation-delay: 0.8s;animation-delay: 0.8s;-ms-animation-delay: 0.8s;-moz-animation-delay: 0.8s;}
span.l-3 {-webkit-animation-delay: 0.6s;animation-delay: 0.6s;-ms-animation-delay: 0.6s;-moz-animation-delay: 0.6s;}
span.l-4 {-webkit-animation-delay: 0.4s;animation-delay: 0.4s;-ms-animation-delay: 0.4s;-moz-animation-delay: 0.4s;}
span.l-5 {-webkit-animation-delay: 0.2s;animation-delay: 0.2s;-ms-animation-delay: 0.2s;-moz-animation-delay: 0.2s;}
span.l-6 {-webkit-animation-delay: 0;animation-delay: 0;-ms-animation-delay: 0;-moz-animation-delay: 0;}

@-webkit-keyframes loader {
	0% {-webkit-transform: translateX(-30px); opacity: 0;}
	25% {opacity: 1;}
	50% {-webkit-transform: translateX(30px); opacity: 0;}
	100% {opacity: 0;}
}

@-moz-keyframes loader {
	0% {-moz-transform: translateX(-30px); opacity: 0;}
	25% {opacity: 1;}
	50% {-moz-transform: translateX(30px); opacity: 0;}
	100% {opacity: 0;}
}

@-keyframes loader {
	0% {-transform: translateX(-30px); opacity: 0;}
	25% {opacity: 1;}
	50% {-transform: translateX(30px); opacity: 0;}
	100% {opacity: 0;}
}

@-ms-keyframes loader {
	0% {-ms-transform: translateX(-30px); opacity: 0;}
	25% {opacity: 1;}
	50% {-ms-transform: translateX(30px); opacity: 0;}
	100% {opacity: 0;}
}
</pre>

`````

```html
<pre class="am-pre-scrollable">
  ...
</pre>
```

## Links:

### Light weight code highlight plugins:

- [google-code-prettify](https://code.google.com/p/google-code-prettify/)
- [highlight.js](https://highlightjs.org/)
- [Rainbow](http://craig.is/making/rainbows)
