package errmsg

var msgMap = map[int]string{
	CodeSuccess:          "成功",
	CodeError:            "服务器内部错误",
	UserNotExist:         "用户不存在",
	UserPasswordError:    "密码错误",
	UserTokenNotExist:    "Token 不存在",
	UserTokenInvalid:     "Token 无效",
	UserTokenExpired:     "Token 已过期",
	UserPermissionDenied: "权限不足",
	UserAccountDisabled:  "用户已被禁用",

	EmailFormatError:     "邮箱格式不正确，请输入正确的 QQ 邮箱",
	EmailLimit60sError:   "请 60 秒后再试",
	EmailLimit1hError:    "1 小时内发送次数过多",
	EmailLimit24hError:   "今日发送次数已达上限",
	EmailSendFailedError: "验证码发送失败，请稍后重试",

	UserAlreadyExistsError: "该用户或邮箱已存在",
	UserUsernameError:      "用户名格式错误：支持中英文下划线，且不超过 10 个字符",
	UserPasswordFormatErr:  "密码格式错误：8-20 位，需包含字母、数字、特殊符号中的两种",
	UserCodeError:          "验证码错误",
	UserCodeNotExist:       "验证码不存在或已过期",
	UserCodeLimitError:     "验证码错误次数过多，请重新获取",
	UserSignatureError:     "个性签名格式错误：支持中英文数字及常用写作符号，且不超过 30 个字",

	LoginLimit1mError:  "登录失败次数过多，请 1 分钟后再试",
	LoginLimit1hError:  "登录失败次数过多，请 1 小时后再试",
	LoginLimit24hError: "今日登录失败次数已达上限，请明日再试",
}

func GetMsg(code int) string {
	if msg, ok := msgMap[code]; ok {
		return msg
	}
	return msgMap[CodeError]
}
