# 常见问题
---

## 使用问题

### 使用时遇到问题如何准确定位？

推荐使用 Chrome 开发者工具：

- [Chrome 开发者工具官方文档](https://developer.chrome.com/devtools)
- [使用 Chrome 开发者工具进行 JavaScript 问题定位与调试](http://www.ibm.com/developerworks/cn/web/1410_wangcy_chromejs/)
- [Google Chrome浏览器开发工具详细教程](http://blog.sina.com.cn/s/blog_6e637ea701017glv.html)

也可以 Firebug 等工具。

### 免费吗？

MIT License，写得很清楚。至于看见广告什么的，都是浮云……

### Amaze UI 目前没有 xx 组件，有没有推荐的？

首先，**确保查看了每个栏目的下面的左侧菜单的每个链接**。如果你的窗口比较小， **左侧的菜单是可以向下滚动**，点小三角或者使用鼠标滚轮上下滚动。

**真的没找到？**那真是没有，参见[插件收集](/getting-started/collections?_ver=2.x)

### 「爱上一匹野马，可你的家里没有草原…」？

Amaze UI 2.x 依赖 jQuery，请在 Amaze UI 的 JS 文件之前引入 jQuery（推荐使用最新版）。

### 拷贝页面中的示例代码粘贴以后没有效果？

- 确保页面中已经引入了 jQuery 以及 Amaze UI 的 CSS、JS，`path/to/` 替换相应的路径。

```html
<link rel="stylesheet" href="path/to/css/amazeui.min.css"/>
<script src="http://code.jquery.com/jquery-2.1.3.min.js"></script>
<script src="path/to/js/amazeui.min.js"></script>
```

- 如果已经引入，请查看 Amaze UI CSS、JS 文件顶部的版本信息，确保引入的版本和示例要求的版本匹配。

### JS 动态插入的 DOM 事件失效？

很多用户遇到[各种「事件失效」问题](https://github.com/allmobilize/amazeui/issues?utf8=%E2%9C%93&q=label%3Aevent-binding+)，在这里做一个统一回答。

> Amaze UI 一般在 jQuery [ready](https://api.jquery.com/ready/) 事件里面初始化默认接口，动态插入的 DOM 已经是在 `ready` 事件之后，自然不会绑定相应事件。

> 一些插件通过**事件委托**可以解决动态插入 DOM 的事件绑定问题，但并不是所有插件都可以这样，有的用户操作没有定位到要绑定事件的 DOM 上，比如窗口滚动。

Amaze UI jQuery 安装传统 website 的思路开发，也就是假定 JS 执行之前 DOM 已经渲染完成。如果是通过 Ajax 或者其他方式动态插入的 DOM，就需要手动调用相关接口初始化插件。

- **JS 插件**

  假设动态插入了一个图片轮播，可以在插入完成以后执行下面的代码初始化：

  ```js
  $('#my-slider').flexslider();
  ```
- **Web 组件**

  有交互行为的 Web 组件提供了初始化的接口，假如动态插入了一个 Figure 组件，可以执行以下代码进行初始化：

  ```js
  AMUI.figure.init();
  ```

更多用户遇到并提出的实际问题，请[点此查看](https://github.com/allmobilize/amazeui/issues?utf8=%E2%9C%93&q=label%3Aevent-binding+)。


