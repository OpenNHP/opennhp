# 开始使用 Amaze UI

---


Amaze UI 是一个轻量级、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写。


## 下载

__注意：__ 目前提供下载的为测试版本，部分细节还在调整、改进，欢迎大家提出意见、建议。

<div class="am-g">
  <div class="col-md-6 col-md-centered">
    <a href="/download" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
">下载 Amaze UI Boilerplate</a>
  </div>
</div>

## jQuery or Zepto?

> 我承认，我是猴子派来捣乱的！

移动端首选 Zepto，桌面端选 jQuery，这应该是大多数开发者的共识。那对于跨平台的响应式网站呢？

- Zepto 体积小，下载快，但 __除了小，还有别的吗？__ Wifi 普及，4G 降临，那几十 KB 的还那么重要吗？优化一张图片好几个 jQuery 就出来了。
- jQuery 体积稍大，这是缺点。但是背后 jQuery 很多细节处理得很到位；成熟的生态圈，很多 jQuery 插件；庞大的社区，使用 jQuery 遇到问题时，可以很快从社区获得解决方案。jQuery 的这些特点有助于有效的提高开发效率。这些都是 Zepto 所缺乏的。
- 性能考量：体积小不等于执行效率高；而且通过数十万次计算得出一个百分之几的差距，实际是放大了性能差异，实际使用中很少有那么大的计算量。

虽然我们现在使用 Zepto，是从专门针对移动开发时代沿袭过来的。现在增加桌面端支持，Zepto 可能[不是一个好的选择](http://zurb.com/article/1293/why-we-dropped-zepto)。

我个人倾向 jQuery，你呢？ __欢迎大家投票，并在[评论](#ds-thread)中分享你的想法__。

<iframe seamless="seamless" style="border: none; overflow: hidden;" height="450" width="100%" scrolling="no" src="http://assets-polarb-com.a.ssl.fastly.net/api/v4/publishers/hegfirose/embedded_polls/iframe?poll_id=192386"></iframe>

## 目录结构

- 以 `index.html` 为样板编写 HTML；
- 在 `app.css` 中编写 CSS；
- 在 `app.js` 中编写 JavaScript；
- 图片资源可以放在 `i` 目录下。

__注意：__ [组件](/widgets) 部分的 CSS 及 JS 已经默认打包，目前只能自行拷贝 HTML 组建页面，稍后会提供基于 [Handlebars](http://handlebarsjs.com/) partial 的更方便的使用方式。


```
.Amaze UI Boilerplate
|
├── assets
│   ├── css
│   │   ├── amui.all.css
│   │   ├── amui.all.min.css
│   │   └── app.css
│   ├── i
│   │   └── favicon.png
│   └── js
│       ├── amui.js
│       ├── amui.min.js
│       ├── app.js
│       └── zepto.min.js
└── index.html
```

## 桌面图标

在开发网站的过程中我们常常会设置一个 Touch Icon。由于系统的差异性，添加到桌面图标并没有统一的标准。iOS 最早支持 Touch Icon，并有明确的规范，其他系统一定程度上支持 iOS 的规范。

### 终极方案

下面是兼容 iOS 4.2+ 及 Android 2.1+ 的通用写法：

```html
<link rel="apple-touch-icon-precomposed" href="/path/to/icon72x72@2x.png">
```
- `rel="apple-touch-icon-precomposed"`：不给 Icon 添加额外的效果；兼容 Android 1.5 - 2.1。
- Icon 尺寸：144px * 144px，兼容 iPhone、iPad 及绝大部分安卓设备。

如果想了解更多细节，可以继续阅读后面的内容。

### iOS

```html
<!-- Add to homescreen for Safari on iOS -->
<!-- 添加至主屏, 从主屏进入会隐藏地址栏和状态栏, 全屏(content="yes") -->
<meta name="apple-mobile-web-app-capable" content="yes">

<!-- 系统顶栏的颜色(content = black、white 和 black-translucent)选其一就可以 -->
<meta name="apple-mobile-web-app-status-bar-style" content="black">
<meta name="apple-mobile-web-app-status-bar-style" content="white">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">

<!-- 指定标题 -->
<meta name="apple-mobile-web-app-title" content="Web Starter Kit">

<!-- 指定icon, 建议PNG格式-->
<link rel="apple-touch-icon" href="touch-icon-iphone.png">
<link rel="apple-touch-icon" sizes="76x76" href="touch-icon-ipad.png">
<link rel="apple-touch-icon" sizes="120x120" href="touch-icon-iphone-retina.png">
<link rel="apple-touch-icon" sizes="152x152" href="touch-icon-ipad-retina.png">

<!--
 使用rel="apple-touch-icon"属性为“增加高光光亮的图标”, 系统会自动为图标添加圆角及高光；
 使用rel="apple-touch-icon-precomposed"属性为“设计原图图标”；
-->
```
尺寸说明：

<table class="am-table am-table-bd am-table-striped" style="text-align: center;">
    <thead style="text-align: center;">
        <tr>
            <th style="width:100px; text-align: center">机型</th>
            <th>iPhone 5 and iPod touch (高分辨率)</th>
            <th>iPhone and iPod touch (高分辨率)</th>
            <th>iPad and iPad mini (高分辨率)</th>
            <th>iPad 2 and iPad mini (标准分辨率)</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>尺寸 (px)</td>
            <td>120 x 120</td>
            <td>120 x 120</td>
            <td>152 x 152</td>
            <td>76 x 76</td>
        </tr>
    </tbody>
</table>

__参考链接：__

- [iOS Human Interface Guidelines - Icon and Image Sizes
](https://developer.apple.com/library/ios/documentation/UserExperience/Conceptual/MobileHIG/IconMatrix.html)
- [Everything you always wanted to know about touch icons](http://mathiasbynens.be/notes/touch-icons)
- [Configuring Web Applications](https://developer.apple.com/library/ios/documentation/AppleApplications/Reference/SafariWebContent/ConfiguringWebApplications/ConfiguringWebApplications.html)

### Android Chrome

Android 下目前只有 Chrome 31 支持 Add to Homescreen。

```html
<!-- Add to homescreen for Chrome on Android -->
<meta name="mobile-web-app-capable" content="yes">
<link rel="icon" sizes="196x196" href="images/touch/chrome-touch-icon-196x196.png">
```

__参考链接：__

- [Chrome Add to Homescreen](https://developer.chrome.com/multidevice/android/installtohomescreen)

### Win 8

```html
<!-- Tile icon for Win8 (144x144 + tile color) -->
<!-- win 8 磁贴图标 -->
<meta name="msapplication-TileImage" content="images/touch/ms-touch-icon-144x144-precomposed.png">
<!-- win 8 磁贴颜色 -->
<meta name="msapplication-TileColor" content="#3372DF">
```

__参考链接：__

- [Pinned Sites][pinedsites]
- [MSDN - Pinned site metadata reference][msdn-pin]

[pinedsites]:http://msdn.microsoft.com/en-us/library/ie/hh772707(v=vs.85).aspx
[msdn-pin]:http://msdn.microsoft.com/zh-cn/library/ie/dn255024(v=vs.85).aspx

## 参与讨论

有任何使用问题，请大家直接在评论中留言，也欢迎大家发表意见、建议。

__感谢大家对 Amaze UI 的关注和支持！__
