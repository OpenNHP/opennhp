# Web 组件简介
---

[Web Components](http://www.w3.org/TR/components-intro/) 颇令人向往，无奈浏览器支持有限，所以，Amaze UI Web 组件按照 Web Components 的实现形式，使用浏览器支持更为普及的技术，将移动开发中常用的组件切割成不同部分，实现类 Web Components 的组件，提高开发效率。

## 组件结构

<div>
  <img src="/i/docs/widget.jpg" alt="Widget 结构" style="max-width: 400px" class="am-center"/>
</div>

如上图所示，Amaze UI Web 组件 通过一个 `package.json` 文件描述，每个组件由模板（hbs）、样式（LESS）、交互（JS）三部分组成，其中样式可能有多个文件（不同的主题）。组件样式和交互以 [CSS](/css) 、[JS 插件](/javascript) 为基础编写；使用 [Handlebars](http://handlebarsjs.com/) 作为模板引擎。

Amaze UI 目前封装的组件及演示请查看 [Web 组件](/widgets) 页。

## 分享组件

如果你想分享你开发的 Web 组件，可以 [Fork Amaze UI 项目](https://github.com/allmobilize/amazeui/fork)，按照[开发文档](/getting-started/widget)开发完成以后，向我们提交 Pull Request。

通过审核以后，你的组件便会出现在 Amaze UI 官网。

欢迎大家加入 Web 组件开发者行列（[开发文档](/getting-started/widget)），为用户开发更多的组件。