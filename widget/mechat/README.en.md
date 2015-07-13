# MeChat
---

[MeChat](http://mechat.im/) online customer service widget.

## Usage

This widget requires a MeChat account and the account ID.

### I don't have an account：

#### Register

Fill the form and send it to the register interface. Your information will be sent to MeChat.

#### After Submission

If the register succeed, your configuration will be sent back.

### I already have an account:

#### Register

Fill the form and send it to the register interface. Your information will be sent to MeChat.

#### After Submission

MeChat will sent your configuration back .Allmobilize will render the ask for your password.

#### Submit again

Your password will be sent to MeChat and your configuration will be sent back.


## API

API is provided by MeChat

1: Register Interface: http://open.mechatim.com/cgi-bin/create/unit2?appid=T4v1KpVM7QOvzxgbQ9

	function: Register the unregistered email and send back configurations.
		      Ask for password if email has been registered.
	Parameters：{
		unitname: Name of company,
		email:  Email of company
	}
	Return: {
		errcode: '0', --Success
		unitid: Your ID,
		form: {
			url: '----',
			type: 'POST',
			desc: '',
			fields: {
				
			}
		}
	}

2：Check Interface：http://open.mechatim.com/cgi-bin/check/unit2?appid=T4v1KpVM7QOvzxgbQ9

	Function: Check password. 
	Parameters:{
		email: 
		password: 
	}
	返回：{
		errcode: '0', --Success
		unitid: Your ID,
		form: {
			url: '----',
			type: 'POST',
			desc: '',
			fields: {

			}
		}
	}

