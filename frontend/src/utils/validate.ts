// Email: 合法 QQ 邮箱
export const validateEmail = (email: string) => {
  const emailRegex = /^[1-9][0-9]{4,10}@qq\.com$/
  if (!emailRegex.test(email)) {
    return '请输入合法的 QQ 邮箱格式'
  }
  return null
}

// Password: 8-20位，包含字母、数字、特殊符号中的至少两种
export const validatePassword = (password: string) => {
  const passRegex = /^(?![0-9]+$)(?![a-zA-Z]+$)(?![!@#$%^&*]+$)[0-9A-Za-z!@#$%^&*]{8,20}$/
  if (!passRegex.test(password)) {
    return '密码需为8-20位，且包含字母/数字/符号中的至少两种'
  }
  return null
}

// Username: 1-10位，中英文、下划线
export const validateUsername = (username: string) => {
  const userRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]{1,10}$/
  if (!userRegex.test(username)) {
    return '用户名需为1-10位中英文或下划线'
  }
  return null
}

// Code: 4位数字或字母
export const validateCode = (code: string) => {
  const codeRegex = /^[a-zA-Z0-9]{4}$/
  if (!codeRegex.test(code)) {
    return '验证码需为4位数字或字母'
  }
  return null
}
