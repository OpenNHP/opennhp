# JavaScript
---

## 关于前端 JS 模块化

> 不好意思，我又来捣乱了。不过有些问题还是想和大家讨论一下。

由于一些历史原因，Amaze UI 目前使用 Sea.js 来组织 JS 模块，但用得并不是很顺，比如在依赖提取、文件合并等环节。当然这可能是因为项目结构、以及最终面向用户的原因。

不过还是想看看大家对前端 JS 模块化的选择，也欢迎大家发表自己的看法。

<script type="text/javascript" charset="utf-8" src="http://static.polldaddy.com/p/8231710.js"></script>
<noscript><a href="http://polldaddy.com/poll/8231710/">浏览器端 JavaScript 模块化，你们的选择是？</a></noscript>

个人倾向于兼容多个标准，把选择权交给用户，比如 FastClick 的写法：

```javascript
if (typeof define == 'function' && typeof define.amd == 'object' && define.amd) {
	// AMD. Register as an anonymous module.
	define(function() {
		'use strict';
		return FastClick;
	});
} else if (typeof module !== 'undefined' && module.exports) {
	module.exports = FastClick.attach;
	module.exports.FastClick = FastClick;
} else {
	window.FastClick = FastClick;
}
```

__建议大家阅读的文章__

* [前端模块化开发那点历史](https://github.com/seajs/seajs/issues/588)
* [Writing Modular JavaScript With AMD, CommonJS & ES Harmony](http://addyosmani.com/writing-modular-js/)

## 初级使用

### 基于 Zepto.js

Amaze UI JavaScript 组件基于 [Zepto.js](http://zeptojs.com/) 开发，使用时确保在 Amaze UI 的脚本之前引入了 Zepto.js（1.1.3）。Zepto.js 的更多细节请查看[官方文档](http://zeptojs.com/)。

由于模块内部指定了 `$ = window.Zepto`，目前并不支持使用 jQuery 替换 Zepto.js，后续的工作中会增加 jQuery 支持。

### 组件调用

组件的调用方式和 jQuery 类似，具体细节请查看各个组件的文档。

## 高级使用

### 基于 Sea.js

Amaze UI 目前使用 [Sea.js](http://seajs.org/) 组织、管理模块，使用 Sea.js 的用户可以通过源码查看相关接口。

### 默认事件接口

Amaze UI 通过特定的 HTML 来绑定默认的事件，多数 JS 组件通过 HTML 标记就可以实现调用。这些默认事件都在 `amui` 命名空间下，用户可以自行关闭。

关闭所有默认事件：

```javascript
$(document).off('.amui');
```

关闭特定组件的默认事件：

```javascript
$(document).off('.modal.amui');
```

### 自定义事件

多数组件都定义了一些自定义事件。

自定义事件命名的方式为 `{事件名称}:{组件名称}:amui`，用户可以查看组件文档使用这些自定义事件。

```javascript
$('#myAlert').on('close:alert:amui', function() {
  // do something
});
```

