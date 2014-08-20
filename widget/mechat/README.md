# MeChat
---

[美洽](http://mechat.im/)在线客服模块（右下角的咨询图标）。

## 使用说明

使用本模块需注册美洽账号并获取帐号ID信息，具体如下。

### 没有美洽帐号的用户：

#### 注册用户

用户填写表单并提交，云适配通过接口发送相关数据到美洽平台。

#### 提交返回

美洽注册成功，返回相应配置，云适配渲染界面，提示添加成功。


### 已有美洽帐号的用户：

#### 注册用户

用户填写表单并提交，云适配通过接口发送相关数据到美洽平台。

#### 提交返回

美洽返回相应配置，云适配渲染界面，提示用户输入相应帐号密码。

#### 再次提交

用户提交，云适配通过接口发送相关数据到美洽平台，美洽返回相应配置，
云适配渲染界面，提示绑定成功。


## API

API由美洽提供。

1：注册接口：http://open.mechatim.com/cgi-bin/create/unit2?appid=T4v1KpVM7QOvzxgbQ9

	功能：邮箱未注册则注册，返回已添加页面配置。
		  已注册则返回密码验证页面配置。
	参数：{
		unitname: 企业名字，
		email:  企业邮箱
	}
	返回：{
		errcode: '0' --表示成功，
		unitid: 用于生成植入网站JS，
		form: {
			url: '----',
			type: 'POST',
			desc: '',
			fields: {

			}
		}
	}

2：验证接口：http://open.mechatim.com/cgi-bin/check/unit2?appid=T4v1KpVM7QOvzxgbQ9

	功能：验证密码，正确则返回绑定成功页面配置。
	参数：{
		email:  企业邮箱,
		password: 密码
	}
	返回：{
		errcode: '0' --表示成功，
		unitid: 用于生成植入网站JS，
		form: {
			url: '----',
			type: 'POST',
			desc: '',
			fields: {

			}
		}
	}

