# 开始使用 Amaze UI
---

Amaze UI 是一个轻量级、 [**Mobile first**](http://cbrac.co/113eY5h) 的前端框架，
基于开源社区流行前端框架编写（[使用、参考的项目列表](https://github.com/allmobilize/amazeui#%E5%8F%82%E8%80%83%E4%BD%BF%E7%94%A8%E7%9A%84%E5%BC%80%E6%BA%90%E9%A1%B9%E7%9B%AE)）。


## 下载

__注意：__ 目前提供下载的为测试版本，部分细节还在调整、改进，欢迎大家提出意见、建议。

<div class="am-g">
  <div class="col-md-6 col-md-centered">
    <a href="/download" class="am-btn am-btn-block am-btn-success am-btn-lg" onclick="window.ga && ga('send', 'pageview', '/download/AmazeUI.zip');
">下载 Amaze UI v1.0.0 beta</a>
  </div>
</div>

## 版本说明

Amaze UI 遵循 [Semantic Versioning](http://semver.org/lang/zh-CN/) 规范，版本格式采用 `主版本号.次版本号.修订号` 的形式，版本号递增规则如下：

- 主版本号：做了不兼容的API 修改，如整体风格变化、大规模重构等；
- 次版本号：做了向下兼容的功能性新增；
- 修订号：做了向下兼容的问题修正、细节调整等。


## 目录结构

- `index.html` - 空白 HTML 模板；
- `blog.html` - 博客页面模板（[预览](/examples/blog.html)）；
- `landing.html` - Landing Page 模板（[预览](/examples/landing.html)）；
- `login.html` - 登录界面模板（[预览](/examples/login.html)）；
- `sidebar.html` - 带边栏的文章模板（[预览](/examples/sidebar.html)）；
- 在 `app.css` 中编写 CSS；
- 在 `app.js` 中编写 JavaScript；
- 图片资源可以放在 `i` 目录下。

```
AmazeUI
├── assets
│   ├── css
│   │   ├── amui.all.css
│   │   ├── amui.all.min.css
│   │   └── app.css
│   ├── i
│   │   ├── app-icon72x72@2x.png
│   │   └── favicon.png
│   └── js
│       ├── amui.js
│       ├── amui.min.js
│       ├── amui.widget.helper.js
│       ├── amui.widget.helper.min.js
│       ├── app.js
│       └── zepto.min.js
├── blog.html
├── index.html
├── landing.html
├── login.html
├── sidebar.html
└── widget.html
```

## 参与讨论

有任何使用问题，请大家直接在评论中留言，也欢迎大家发表意见、建议。

__感谢大家对 Amaze UI 的关注和支持！__

## jQuery or Zepto?

> 我承认，我是猴子派来捣乱的！

移动端首选 Zepto，桌面端选 jQuery，这应该是大多数开发者的共识。那对于跨平台的响应式网站呢？

- Zepto 体积小，下载快，但 __除了小，还有别的吗？__ Wifi 普及，4G 降临，那几十 KB 的还那么重要吗？优化一张图片好几个 jQuery 就出来了。
- jQuery 体积稍大，这是缺点。但是背后 jQuery 很多细节处理得很到位；成熟的生态圈，很多 jQuery 插件；庞大的社区，使用 jQuery 遇到问题时，可以很快从社区获得解决方案。jQuery 的这些特点有助于有效的提高开发效率。这些都是 Zepto 所缺乏的。
- 性能考量：体积小不等于执行效率高；而且通过数十万次计算得出一个百分之几的差距，实际是放大了性能差异，实际使用中很少有那么大的计算量。

虽然我们现在使用 Zepto，是从专门针对移动开发时代沿袭过来的。现在增加桌面端支持，Zepto 可能[不是一个好的选择](http://zurb.com/article/1293/why-we-dropped-zepto)。

我个人倾向 jQuery，你呢？ __欢迎大家投票，并在[评论](#ds-thread)中分享你的想法__。

<iframe seamless="seamless" style="border: none; overflow: hidden;" height="450" width="100%" scrolling="no" src="http://assets-polarb-com.a.ssl.fastly.net/api/v4/publishers/hegfirose/embedded_polls/iframe?poll_id=192386"></iframe>
