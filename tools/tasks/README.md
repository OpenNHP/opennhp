# Tasks

## customizer.js 定制工具
----------

**定制流程**：

1. **下载源码：**从 GitHub 选择版本，点击 `Source code (zip)` 下载并解压（定制只适用于 Amaze UI 2.x）；
2. 安装 [Node.js](http://nodejs.org/download/)；
3. **全局安装 gulp**：`npm install -g gulp`；
4. **安装依赖**：切换到 Amaze UI 源码目录，运行 `npm install`；
5. **下载定制配置文件**：[选择需要的组件](http://amazeui.org/customize)，然后点击绿色按钮下载配置文件；
6. **定制**：将下载的 `config.json` 放到源码 `tools/tasks/` 下，在源码根目录执行 `gulp customize`，定制的代码放在 `dist/customized/` 下。

**必选的基础依赖（默认勾选）：**

- css: `base` `mixins` `variables`
- js: `core`
