# 助力 Amaze UI

欢迎拾柴，助力 Amaze UI 发展。感谢所有的参与者！

如果你没有相关经验，建议先阅读下面的文章：

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [The Beginner’s Guide to Contributing to Open Source Projects](http://blog.newrelic.com/2014/05/05/open-source_gettingstarted/)
- [How to Start Contributing to Open Source](http://www.developer.com/open/how-to-start-contributing-to-open-source.html)

## Bug 反馈

- 查看 [Issue
系统](https://github.com/allmobilize/amazeui/issues/new?title=Bug%3A%20&body=**%E9%97%AE%E9%A2%98%E6%8F%8F%E8%BF%B0**%0A%0A%EF%BC%88%E6%8F%8F%E8%BF%B0%E4%B8%80%E4%B8%8B%E9%97%AE%E9%A2%98%EF%BC%89%0A%0A**%E4%BA%A7%E7%94%9F%E7%8E%AF%E5%A2%83**%0A%0A-%20%E6%93%8D%E4%BD%9C%E7%B3%BB%E7%BB%9F%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%B5%8F%E8%A7%88%E5%99%A8%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%BC%94%E7%A4%BA%E5%9C%B0%E5%9D%80%EF%BC%9A%0A%0A**%E5%A4%8D%E7%8E%B0%E6%AD%A5%E5%A5%8F**%0A%0A1.%20%0A2.%20%0A...) 或者相关评论，确认 Bug
是否已经有人提交过，如果有，请直接在回复以前的反馈者；
- 确认 Bug 是否已经被修复；
- 隔离使用场景，确认 Bug 是否由第三方文件导致；
- 详细描述 Bug 出现的场景，包括：操作系统、浏览器及版本等；
- 如有可能尽量添加截图；
- 尽量添加可以在线查看的演示页面，推荐使用 JSBin 等工具。

为了方便问题重现，我们使用 **[JSBin 搭建了在线一个调试工具](http://bin.amazeui.org)**，你可以把有问题的场景粘在里面，反馈给我们。

**Bug 反馈模板**（[点击使用模板](https://github.com/allmobilize/amazeui/issues/new?title=Bug%3A%20&body=**%E9%97%AE%E9%A2%98%E6%8F%8F%E8%BF%B0**%0A%0A%EF%BC%88%E6%8F%8F%E8%BF%B0%E4%B8%80%E4%B8%8B%E9%97%AE%E9%A2%98%EF%BC%89%0A%0A**%E4%BA%A7%E7%94%9F%E7%8E%AF%E5%A2%83**%0A%0A-%20%E6%93%8D%E4%BD%9C%E7%B3%BB%E7%BB%9F%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%B5%8F%E8%A7%88%E5%99%A8%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%BC%94%E7%A4%BA%E5%9C%B0%E5%9D%80%EF%BC%9A%0A%0A**%E5%A4%8D%E7%8E%B0%E6%AD%A5%E5%A5%8F**%0A%0A1.%20%0A2.%20%0A...)）

```
**问题描述**

（描述一下问题）

**产生环境**

- 操作系统及版本：
- 浏览器及版本：
- 演示地址：

**复现步奏**

1.
2.
...
```

## 功能需求

欢迎提交功能需求。不过提交之前先看看以下几点：

- 访问[官方网站](http://amazeui.org/)，浏览相应栏目，确定是否已经有你需要的功能；
- 查看 [Issue
系统](https://github.com/allmobilize/amazeui/issues)及[开发路线图](https://github.com/allmobilize/amazeui/wiki/Roadmap)，确定需求是否已经有人提交过，或者已经在官方的计划列表中；
- 需求必须符合我们的目标和理念（面向现代浏览器）；
- 尽可能详细描述需求，如果能附上实例更加。

## Pull Request

Pull Request 之前，请先阅读 [Amaze UI 编码规范](https://github.com/allmobilize/amazeui/wiki/Style-Guide)。

（相关代码检测、测试流程陆续完善中。）


## 我们希望用户参与的项目

- 把[使用 Amaze UI 的项目](https://github.com/allmobilize/amazeui/wiki/Sites-Using-AmazeUI)提交给我们，能获得一定的展示量，同时能给其他用户提供参考；
- SCSS 等其他 CSS 扩展语言的支持：通过编写自动转换脚本实现，人工迁移同步更新成本太高；
- 使用 Amaze UI 开发制作模板；
- Angular.js、Ember.js、Meteor 等框架的支持；
- 基于 Amaze UI 样式开发 JavaScript 插件，可以把 Bootstrap 的插件移植到 Amaze UI；
- 兼容性测试，我们的测试只能覆盖有限，需要更多用户参与，包括浏览器、WebView 等不同的用户代理；
- ...

## 开发文档

开发文档存放在 `docs/styleguide` 目录下，也可以通过 [Amaze UI 官网](http://amazeui.org/)查看：

- [Amaze UI HTML/CSS 编写规范](http://amazeui.org/getting-started/html-css)
- [Amaze UI JavaScript 编写规范](http://amazeui.org/getting-started/javascript)
- [Amaze UI Web 组件开发规范](http://amazeui.org/getting-started/widget)
