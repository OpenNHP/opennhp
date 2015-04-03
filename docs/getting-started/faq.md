# 常见问题
---

## 版本更新

### 下一阶段会做什么?

与前端 MV* 框架整合，深挖 SPA。


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

