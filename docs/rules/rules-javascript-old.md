# Javascript 开发规范

## 命名规范

1. 文件和目录名只能包含 [a-z\d\-]，并以英文字母开头
2. 首选合适的英文单词 
3. data api 命名为小写并用连字符，如 data-trigger-type
4. 事件名为驼峰，如 .trigger('itemSelected')
5. 符合规范
   - 常量全大写 UPPERCASE_WORD
   - 变量驼峰 camelName
   - 类名驼峰，并且首字母要大写 CamelName

## 目录结构

```
dist //全局完整构造打包后内容
  css
    amui.css
    amui.min.css
    amui-theme.css
    amui-theme.min.css
  js
    amui.js
    amui.min.js
  fonts
    ...

docs     //全局说明文档，演示网站，规范文档 markdown格式
vendor   //第三方依赖类库 靠bower自动更新，不上传到github
js       //全局依赖的js（core,utils）
less     //全局依赖的less(reset,core)
resource //全局依赖的静态资源（主要是图标）
fonts    //公共使用的图标(暂不实现)
 
widget //模块组件 
  widget1    //模块子目录结构
    dist     //模块打包后内容
    examples //模块组件的实例
    docs     //模块的说明文档
    src      //源代码
      *.js   //基于sea.js,靠相对路径依赖基础js
      *.less //基于@import，靠相对路径依赖基础less
      *.hbs  //基于handlberbars语法的html模板
    HISTORY.md //历史修改记录
    README.md  
    package.json //模块的配置和依赖
  widget2
  ......

```

## 编码风格

需要通过 JSLint，查看具体[编码风格](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E7%BC%96%E7%A0%81%E9%A3%8E%E6%A0%BC)

## 注释规范

不建议使用 jsdoc，注释的目的是：**提高代码的可读性，从而提高代码的可维护性。**

查看具体[注释规范](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E6%B3%A8%E9%87%8A%E8%A7%84%E8%8C%83)

## 文档规范

### README.md

每个组件必须有 README.md 文件，用来描述组件的基本情况。

```
# 模块名称

-----

该模块的概要介绍。

------

## 使用说明

如何使用该模块，可以根据组件的具体特征，合理组织。

## API

需要提供 API 说明，属性、方法、事件等。
```

### docs

如果组件需要写的东西比较多，可以划分好放到 docs 下。比如竞争者分析，多模块的情况。

### HISTORY.md

记录组件的变更，最好和 issues 进行绑定。请阅读[历史记录书写规范](https://github.com/aralejs/aralejs.org/wiki/%E5%8E%86%E5%8F%B2%E8%AE%B0%E5%BD%95%E4%B9%A6%E5%86%99%E8%A7%84%E8%8C%83)。

```
### 1.1.0

* [tag:fixed] #18 修复了 XXX 问题
* [tag:fixed] #29 修复了 YYY 问题
* [tag:new] #12 增加了 ZZZ 功能
* [tag:improved] #23 优化了 BBB 代码

### 1.0.0

* [tag:new] 第一个发布版本
```


## Reference

 -  [注释规范](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E6%B3%A8%E9%87%8A%E8%A7%84%E8%8C%83)

 -  [编码风格](https://github.com/aralejs/aralejs.org/wiki/JavaScript-%E7%BC%96%E7%A0%81%E9%A3%8E%E6%A0%BC)

 -  [编码与文档的讨论](https://github.com/aralejs/aralejs.org/issues/36)

 -  [常用词命名统一表](https://github.com/aralejs/aralejs.org/wiki/%E5%B8%B8%E7%94%A8%E8%AF%8D%E5%91%BD%E5%90%8D%E7%BB%9F%E4%B8%80%E8%A1%A8)
