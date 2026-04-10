package controller

import (
	"backend/dto"
	"backend/pkg/errmsg"
	"backend/pkg/utils"
	"backend/pkg/utils/response"
	"backend/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
	authService service.AuthService
}

// SendRegisterEmail 发送注册验证码接口
func (u *UserController) SendRegisterEmail(c *gin.Context) {
	var req dto.SendEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	// 校验邮箱格式
	if !utils.IsQQEmail(req.Email) {
		response.FailWithCode(errmsg.EmailFormatError, c)
		return
	}

	// 调用服务发送验证码
	code := u.userService.SendRegisterCode(c.Request.Context(), req.Email)
	if code != errmsg.CodeSuccess {
		response.FailWithCode(code, c)
		return
	}

	response.Ok(nil, c)
}

// Register 用户注册接口
func (u *UserController) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	// 1. 格式校验
	if !utils.IsUsername(req.Username) {
		response.FailWithCode(errmsg.UserUsernameError, c)
		return
	}
	if !utils.IsQQEmail(req.Email) {
		response.FailWithCode(errmsg.EmailFormatError, c)
		return
	}
	if !utils.IsPassword(req.Password) {
		response.FailWithCode(errmsg.UserPasswordFormatErr, c)
		return
	}
	if !utils.IsCode(req.Code) {
		response.FailWithCode(errmsg.UserCodeError, c)
		return
	}
	if req.Signature != "" && !utils.IsSignature(req.Signature) {
		response.FailWithCode(errmsg.UserSignatureError, c)
		return
	}

	// 2. 调用服务执行注册
	errCode, token, refreshToken, expiresAt, userInfo := u.userService.Register(c.Request.Context(), req.Username, req.Email, req.Password, req.Code, req.Signature)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(dto.RegisterRes{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         userInfo,
	}, c)
}

// Login 用户登录接口
func (u *UserController) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	// 1. 格式校验
	if !utils.IsQQEmail(req.Email) {
		response.FailWithCode(errmsg.EmailFormatError, c)
		return
	}
	if !utils.IsPassword(req.Password) {
		response.FailWithCode(errmsg.UserPasswordFormatErr, c)
		return
	}

	// 2. 调用服务执行登录
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	errCode, token, refreshToken, expiresAt, userInfo := u.userService.Login(c.Request.Context(), req.Email, req.Password, ip, ua)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(dto.LoginRes{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         userInfo,
	}, c)
}

// GetUserInfo 获取用户信息接口
func (u *UserController) GetUserInfo(c *gin.Context) {
	// 从 context/header 中获取 userId
	userId, err := u.authService.GetUserID(c)
	if err != nil {
		response.FailWithCode(errmsg.UserTokenInvalid, c)
		return
	}

	errCode, userInfo := u.userService.GetUserInfo(c.Request.Context(), userId)
	if errCode != errmsg.CodeSuccess {
		response.FailWithCode(errCode, c)
		return
	}

	response.Ok(userInfo, c)
}

// RefreshToken 刷新 Token 接口
func (u *UserController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(errmsg.CodeError, c)
		return
	}

	token, expiresAt, err := u.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		// 如果是刷新 Token 过期或者无效，返回专门的错误码
		response.FailWithCode(errmsg.UserTokenExpired, c)
		return
	}

	response.Ok(dto.RefreshTokenRes{
		Token:     token,
		ExpiresAt: expiresAt,
	}, c)
}
