# Bug 反馈
---

感谢对 Amaze UI 的关注和支持，如遇到 Bug 或者使用问题，可以通过以下途径反馈给我们：

- 在组件文档底部[评论中留言](#ds-thread)；
- 在 GitHub 项目主页<a class="new-issue" href="https://github.com/allmobilize/amazeui/issues/new?title=Bug%3A%20&body=**%E9%97%AE%E9%A2%98%E6%8F%8F%E8%BF%B0**%0A%0A%EF%BC%88%E6%8F%8F%E8%BF%B0%E4%B8%80%E4%B8%8B%E9%97%AE%E9%A2%98%EF%BC%89%0A%0A**%E4%BA%A7%E7%94%9F%E7%8E%AF%E5%A2%83**%0A%0A-%20%E8%AE%BE%E5%A4%87%EF%BC%9A%EF%BC%88%E6%89%8B%E6%9C%BA%E3%80%81%E5%B9%B3%E6%9D%BF%E7%AD%89%E7%A7%BB%E5%8A%A8%E8%AE%BE%E5%A4%87%E6%97%B6%E5%A1%AB%E5%86%99%E6%AD%A4%E9%A1%B9%EF%BC%89%0A-%20%E6%93%8D%E4%BD%9C%E7%B3%BB%E7%BB%9F%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%B5%8F%E8%A7%88%E5%99%A8%E5%8F%8A%E7%89%88%E6%9C%AC%EF%BC%9A%0A-%20%E6%BC%94%E7%A4%BA%E5%9C%B0%E5%9D%80%EF%BC%9A%0A%0A**%E5%A4%8D%E7%8E%B0%E6%AD%A5%E5%A5%8F**%0A%0A1.%20%0A2.%20%0A...%20%0A%20%0A**UA:**%20%0A">提交 Issue</a>。

<div class="am-alert am-alert-danger">
  我们感谢提交的 Bug 的同学，但也请提交的时候换位思考一下：如果别人给你提交一个这样的 Issue，你能快速准确的理解吗？如果不能，烦请重新整理你的语言，按照要求的格式填写。专业一点，减少不必要的口舌浪费。
</div>

## Bug 反馈要求

- 标题：简要描述 Bug；
- 内容：
  - 描述一下 Bug，以及 Bug 产生的环境（操作系统及版本，浏览器以及版本）；
  - 如有可能，描述 Bug 复现的流程；
  - 如有可能，添加产生 Bug 时的截图；
  - 尽量添加 Bug 测试的 URL，推荐使用 [JSBin](http://jsbin.com/?html,output)。

```
**问题描述**

（描述一下问题）

**产生环境**

- 设备：（手机、平板等移动设备时填写此项）
- 操作系统及版本：
- 浏览器及版本：
- 演示地址：

**复现步奏**

1.
2.
...
```
<!--
https://github.com/allmobilize/amazeui/issues/new?title=Bug%3A%20&body=**问题描述**%0A%0A（描述一下问题）%0A%0A**产生环境**%0A%0A- 设备：（手机、平板等移动设备时填写此项）%0A- 操作系统及版本：%0A- 浏览器及版本：%0A- 演示地址：%0A%0A**复现步奏**%0A%0A1. %0A2. %0A...-->

## 注意事项

__提交反馈__：

为了能最准确的传达所描述的问题，__建议你在反馈时附上演示__，方便我们理解及更快速的定位、解决问题。

我们很不喜欢重复下面的对话：

```
用户甲：xxx 有问题！

Amaze UI：什么问题？操作系统及版本、浏览器及版本？

用户甲：巴拉巴拉
```

下面的几个链接是我们在几个在线调试工具上建的页面，__已经引入了 Amaze UI 样式和脚本__，你可以<span class="am-text-danger">【Fork】</span>一份，把要有问题的场景粘在里面，反馈给我们。

- [Debug Amaze UI 2.x in JSBin](http://jsbin.com/zoqaba/1/edit?html,output)
- [Debug Amaze UI 1.x in JSBin](http://jsbin.com/qasoxibuje/1/edit?html,output)

__反馈处理__:

提交到 GitHub 的 Issue 一般会通过两种方式关闭：

- __涉及代码修改的问题__：一般会通过 Commit 关闭 Issue，在收到关闭通知以后你可以更新代码确认问题是否已经被修复。如果问题依然存在，劳烦 Reopen Issue 并把问题细节提交给我们；
- __使用问题__：在我们给出答案后，希望你能反馈一下是否解决了你的问题。如果解决了，请关闭 Issue；如果未解决，请描述具体细节。如给出答案一周后仍无任何回复的，Issue 将被关闭。

## 常见问题

### 支持贴图吗？

GitHub Issue 系统支持贴图：

- 把**图片拖到输入框**里图片会自动上传并插入到内容中；
- 也可以**直接复制粘贴图片**。

多说评论系统见下面。

### 如何在多说评论框里面贴图？

~~多说评论系统的贴图功能已经开启，但是多说没有提供图片服务器，仍然需要通过第三方图床来完成。~~

~~点击**插入图片**图标以后，把 `请输入图片地址` 替换成图床里的图片地址。~~

鉴于开启 HTML 解析以后，用户粘贴的代码被解析了。现在禁用了，无法直接插入图片了，不过仍然可以使用下面的图床上传以后给出图片地址。

- [Gimhoy图床](http://pic.gimhoy.com/)
- [围脖图床修复计划（浏览器插件）](http://weibotuchuang.sinaapp.com/)
