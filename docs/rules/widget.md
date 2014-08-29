# Amaze UI Web 组件开发规范
---

Web 组件基于 Amaze UI 基础库（[CSS](/css/) / [JS](/javascript)）开发，在基础库已有样式、功能的基础上做更多扩展。

## Web 组件样式组织

Web 组件的样式有三个层级：

1. Amaze UI 基础样式: 每个网站项目中都会默认引入以下基础样式，组件开发时应在以下样式的基础上进行。
    - `variables.less`
    - `mixins.less`
    - `base.less`
    - `grid.less`
    - `block-grid.less`
    - `utility.less`

2. Amaze UI 其他样式组件：Web 组件编写过程中使用到类似的样式时应当引入相关 [CSS 组件](/css/)，在此上进行微调，比如 `button.less`、`close.less`。

3. Web 组件自身样式：Web 组件自身样式拆分出骨干样式和主题样式；每个 Web 组件可以有多个不同主题，主题基于骨干样式编写，每个主题相互独立。
    - 骨干样式以 `{widget}.less` 命名；
    - 默认主题以 `{widget}.default.less` 命名；
    - 其他主题以 `{widget}.{theme-name}.less` 命名。

## 目录结构及说明

一个组件的目录结构如下：

```
{widget}
|-- package.json
|-- README.md
|-- HISTORY.md
|-- resources
|   `-- screenshots
|       |-- 0.jpg
|       |-- 1.jpg
|       `-- 2.jpg
`-- src
    |-- {widget}.hbs
    |-- {widget}.js
    |-- {widget}.less
    |-- {widget}.default.less
    |-- {widget}.[theme1].less
    |-- {widget}.[themen].less
    `-- {widget}.png
```

### `package.json`

Web 组件核心描述文件，`json` 格式，下面的注释仅为方便解释各项含义添加。

```javascript
{
    // 组件名称: 使用小写字母，不能和存在的组件重名
    "name": "sample",

    // 版本号
    "version": "0.0.1",

    // 组件本地化名字，目前有中文、英文两个选项
    "localName": {
        "en": "",
        "zh-cn": ""
    },

    // 组件类型 [""|"layout"|"social"]
    "type": "",

    // 组件 ICON，存放在 src 目录下
    "icon": "sample.png",

    // 作者信息
    "author": {
        "name": "xxx",
        "email": "xxx@yunshipei.com"
    },

    // 组件描述
    "description": "sample 描述",

    // 组件驱动者
    "poweredBy": "AllMobilize",

    // 基础样式（无需修改）
    "styleBase": [
        "variables.less",
        "mixins.less",
        "base.less",
        "grid.less",
        "block-grid.less",
        "utility.less"
    ],

    // 组件模板，使用 [handlebarsjs](http://handlebarsjs.com/)
    "template": "sample.hbs",

    // 依赖的库样式
    "styleDependencies": [
        "icon.less"
    ],

    // 组件核心样式
    "style": "sample.less",

    // 组件主题（没有主题时将值设置为 null）
    "themes": [
        {
            // 主题名称 sample.{xxx}.less 中的 {xxx}，尽量语义化描述主题
            "name": "default",
            // 主题描述，简要描述主题
            "desc": "默认",
            // 主题使用配置选项
            "options": {},
            // 主题钩子
            "hook": "hook-am-sample-default",
            // 主题使用的less变量
            "variables": [
                {
                    // 变量名
                    "variable": "",
                    // 变量描述名字
                    "name": "",
                    // 默认值
                    "default": "",
                    // 使用改变量的 css 样式
                    "used": [
                        {
                            "selector": "",
                            "property": ""
                        }
                    ]
                }
            ],
            // 主题演示数据，可以为多个
            "demos": [
                {
                    // 演示描述
                    "desc": "",
                    // 演示数据
                    "data": {}
                }
            ]
        }
    ],

    // Amaze UI 核心js（无需修改）
    "jsBase": [
        "core.js"
    ],

    // 依赖的 Amaze UI js 插件
    "jsDependencies": [],

    // 组件脚本
    "script": "sample.js",

    // api 用于生成用户 GUI 界面以及保存用户提交的数据
    "api": {
        "id": {
            "name": "ID", // 表单提示名称
            "desc": "组件自定义ID，遵循CSS ID命名规范",
            "type": "text", // 表单类型
            "default": "", // 默认值
            "pattern": "", // 表单验证正则表达式
            "required": false // 是否为必填
        },
        "className": {
            "name": "Class",
            "desc": "用户自定义组件 class，遵循 CSS class 命名规范",
            "type": "text",
            "default": "",
            "required": false
        },
        // 主题选择（没有主题时将值设置为 null）
        "theme": {
            "name": "主题",
            "desc": "组件主题",
            "type": "select", // 下拉选框
            "default": "default",
            "required": true,
            "dataList": "<%= pkg.themes %>" // 从 themes 中读取主题列表
        },
        // 组件选项（没有选项时将值设置为 null）
        "options": {
            "multiple": {
                "name": "同时展开多个面板",
                "desc": "是否允许同时展开多个面板",
                "type": "select",
                "default": false,
                "required": false,
                // 表单类型为 select 时通过 dataList 设置 <option> 数据
                "dataList": [
                    {
                        "value": "false",
                        "title": "不启用",
                        "selected": true
                    },
                    {
                        "value": true,
                        "title": "启用"
                    }
                ]
            }
        },

        // 内容
        "content": {
        	  // 内容类型 Array 数组、Object（对象）
            "type": "Array",
            "item": {
                "title": {
                    "type": "text",
                    "comment": "面板标题"
                },
                "content": {
                    "type": "html",
                    "comment": "面板内容"
                }
            }
        }

        // 表单接口 - 测试接口，细节还会做调整
        // 提供的表单接口必须支持跨域调用，并返回 JSON 数据
        "forms": {
            "signin": {
                "url": "http://api.xxx.com/signin", // 提交数据的接口
                "type": "POST",
                "desc": "登录表单，你可以在这里写表单的描述信息",
                "fields": { // 表单字段，字段名称应该和返回数据里的字段对应
                    "username": {
                        "name": "用户名",
                        "placeholder": "请使用邮箱",
                        "type": "text",
                        "default": "",
                        "required": true
                    },
                    "password": {
                        "name": "密码",
                        "placeholder": "设置密码",
                        "type": "text",
                        "default": "",
                        "required": true
                    },
                    "submit": {
                        "name": "提交信息",
                        "type": "submit",
                        "default": ""
                    }
                }
            },

            "signup": {
                "url": "http://api.xxx.com/signup",
                "type": "POST",
                "desc": "注册表单，你可以在这里写表单的描述信息",
                "fields": {
                    "username": {
                        "name": "用户名",
                        "placeholder": "请使用邮箱",
                        "type": "text",
                        "default": "",
                        "required": true
                    },
                    "password": {
                        "name": "密码",
                        "placeholder": "设置密码",
                        "type": "text",
                        "default": "",
                        "required": true
                    }
                }
            }
        }
    },

    // 是否隐藏组件 - 此选项供云适配 WebIDE 使用
    "hidden": false
}
```

### `README.md`

Web 组件使用说明，包括 Web 组件 API 介绍、使用技巧、注意事项等。

### `HISTORY.md`

Web 组件更新历史记录。

### `src` 目录

`src` 目录包含 Web 组件的模板（hbs）、核心样式（less）、交互（js）、图标（png）、主题文件。

`src` 目录里的文件使用 Web 组件名称作为文件名，主题使用 `{Web 组件名}.{主题名}.less` 形式命名。

- `{widget}.hbs`: 模板，使用 handlebars
- `{widget}.less`: 主题核心样式，使用 less 编写
- `{widget}.js`: 组件交互
- `{widget}.png`: 组件图标，50px * 50px
- `{widget}.default.less`: 默认主题
- `{widget}.xxx.less`: 其他主题，可以任意多

#### 模板 `{widget}.hbs`

* `data-am-widget="figure"` 为统一标识符；
* `.am-{Web 组件名}` 为 Web 组件基础标识符，Web 组件的所有子元素、主题、状态基于此命名；
  Web 组件基础标识符采用 `am-{widget}` 方式命名，子元素、主题、状态采用 `am-{widget}-{子元素|主题|状态}`，如
`am-figure-hover` 、 `am-figure-bd` 、 `am-figure-active` 、 `am-figure-ios7`。
* 模板使用 `{{#this}}..{{this}}` 包裹（Web 组件使用时注册为 Handlebars `partial`，通过 `{{> widget data}}` 的形式调用。

```html
{{#this}}
  <figure data-am-widget="figure"
      class="am-figure{{#if theme}} am-figure-{{theme}}{{else}} am-figure-default{{/if}}{{#if options.zoomble}} am-figure-zoomable{{/if}}{{#if widgetId}} {{widgetId}}{{/if}}{{#if className}} {{className}}{{/if}}"{{#if id}}
      id="{{id}}"{{/if}}>
    {{#if content.img}}
      <img src="{{content.img}}" alt="{{#if content.imgAlt}}{{content.imgAlt}}{{else}}{{content.figcaption}}{{/if}}"/>
    {{/if}}

    {{#if content.figcaption}}
       <figcaption class="am-figure-capition">
         {{content.figcaption}}
       </figcaption>
    {{/if}}
  </figure>
{{/this}}
```

**注意**：

* Handlebars 模板中不支持 `<script>` 标签，如需加载外部脚本需在 `{widget}.js` 中进行。
* 如果需要保存用户设置的选项，使用 `data-am-{widget}-{option}` 保存在模板中，然后在 `{widget}.js` 中通过 `attr('data-am-{widget}-{option}')` 读取。


#### 核心样式 `{widget}.less`

Web 组件样式使用 less 编写。

```css
.am-{widget} {

  .hook-am-{widget};
}

.hook-am-{widget}() {}
```

样式添加必要的 `hook`，方便用户修改。

#### 默认主题 `{widget}.defalt.less`

```css
.am-{widget}-default {

  .hook-am-{widget}-default;
}

.hook-am-{widget}-default() {}
```

#### 其他主题 `{widget}.xxx.less`

```css
.am-{widget}-xxx {

  .hook-am-{widget}-xxx;
}

.hook-am-{widget}-xxx() {}
```

#### Web 组件交互 `{widget}.js`

Amaze UI 使用 Seajs 、Zepto，Widget 的脚本需按照 Seajs 规范编写。

__如果要在 JS 中动态插入外部样式、脚本，必须在 `load` 事件触发以后再执行相关操作，以免影响网页基本内容载入。__

```javascript
define(function(require, exports, module) {
    // 按此方式使用 Zepto
    var $ = window.Zepto;

});
```

## 开发脚手架

我们提供一个基于 [Slush.js](http://slushjs.github.io/) 的[开发脚手架](https://www.npmjs.org/package/slush-amuiwidget)，可以快速生成 Web 组件目录及相关文件。

全局安装 Slush:

```bash
npm install -g slush
```

全局安装 `slush-amuiwidget`:

```bash
npm install -g slush-amuiwidget
```

在 Amaze UI 项目根目录下面执行:

```bash
slush amuiwidget
```

## 调试预览

按照规范开发完 Web 组件以后，可以在本地调试预览组件。

在 Amaze UI 项目根目录下执行以下命令，安装依赖：

```
npm install
```

全局安装 `nodemon`：

```
npm install nodemon -g
```

安装完成以后执行：

```
gulp preview
```

然后在浏览器里打开 `http://localhost:3008/#{component}` 查看组件的效果，`{component}` 替换为组件名称。

有样式、脚本、配置文件修改时，修改完成以后刷新浏览器即可，`nodemon` 会自动重启 Node 服务。

