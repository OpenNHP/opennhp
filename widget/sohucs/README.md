# Sohucs 搜狐畅言
---

[搜狐畅言](http://changyan.sohu.com/) 社交评论系统模块。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的参数替换为自己的。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `sohucs`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面，设置相关参数；
- 本组件无需采集数据。

## 配置信息获取说明

使用本模块需注册搜狐畅言账号并获取 API Key，具体如下。

### 注册用户

点击链接填写邮箱和密码注册搜狐畅言账号。

![搜狐畅言注册界面](http://changyan.sohu.com/help/images/index-img-01.png)

填写邮箱和密码，点击 `注册` 按钮，出现如下界面：

![填写站点信息](http://amazeui.b0.upaiyun.com/assets/i/cpts/sohucs/sohucs-help-1.png)

输入网站名称、网址、QQ，点击 `提交并获取代码`。

### 获取 API Key

完成上述操作后，进入以下界面，获取 API Key (AppID、Conf)。

![获取代码界面](http://amazeui.b0.upaiyun.com/assets/i/cpts/sohucs/sohucs-help-2.png)

如上图所示，复制 AppID 及 Conf（**红框部分，不包含单引号**），填入云适配 IDE 参数设置里完成配置。

### 注意事项

- 此模块使用搜狐畅言提供的服务，更多使用细节参见[搜狐畅言帮助信息](http://changyan.sohu.com/help/)。


## 数据接口

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {
    "appid": "",
    "conf": ""
  }
}
```
