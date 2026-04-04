package service

import (
	"backend/dto"
	"backend/global"
	"backend/model"
	"backend/pkg/errmsg"
	"backend/pkg/utils"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserService struct{}

// GetUserInfo 获取用户信息 (返回脱敏后的 DTO)
func (u *UserService) GetUserInfo(ctx context.Context, userID uint) (int, dto.UserInfoRes) {
	var userInfo dto.UserInfoRes
	user, err := model.GetUserByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errmsg.UserNotExist, userInfo
		}
		return errmsg.CodeError, userInfo
	}

	userInfo = dto.UserInfoRes{
		Username:  user.Username,
		Email:     utils.DesensitizeEmail(user.Email),
		AvatarUrl: user.AvatarUrl,
		Signature: user.Signature,
	}

	return errmsg.CodeSuccess, userInfo
}

// Login 用户登录
func (u *UserService) Login(ctx context.Context, email, password, ip, ua string) (int, string, string, int64, dto.UserInfoRes) {
	// 1. 检查频率限制 (基于邮箱)
	key1m := fmt.Sprintf("login:fail:%s:1m", email)
	key1h := fmt.Sprintf("login:fail:%s:1h", email)
	key24h := fmt.Sprintf("login:fail:%s:24h", email)

	count1m, _ := global.GVA_REDIS.Get(ctx, key1m).Int()
	if count1m >= 3 {
		return errmsg.LoginLimit1mError, "", "", 0, dto.UserInfoRes{}
	}
	count1h, _ := global.GVA_REDIS.Get(ctx, key1h).Int()
	if count1h >= 5 {
		return errmsg.LoginLimit1hError, "", "", 0, dto.UserInfoRes{}
	}
	count24h, _ := global.GVA_REDIS.Get(ctx, key24h).Int()
	if count24h >= 7 {
		return errmsg.LoginLimit24hError, "", "", 0, dto.UserInfoRes{}
	}

	// 2. 检查用户是否存在
	user, err := model.GetUserByEmail(ctx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errmsg.UserNotExist, "", "", 0, dto.UserInfoRes{}
		}
		return errmsg.CodeError, "", "", 0, dto.UserInfoRes{}
	}

	// 3. 校验密码
	if !utils.BcryptCheck(password, user.PasswordHash) {
		// 密码错误，增加失败记录
		global.GVA_REDIS.Incr(ctx, key1m)
		if count1m == 0 {
			global.GVA_REDIS.Expire(ctx, key1m, 1*time.Minute)
		}
		global.GVA_REDIS.Incr(ctx, key1h)
		if count1h == 0 {
			global.GVA_REDIS.Expire(ctx, key1h, 1*time.Hour)
		}
		global.GVA_REDIS.Incr(ctx, key24h)
		if count24h == 0 {
			global.GVA_REDIS.Expire(ctx, key24h, 24*time.Hour)
		}

		// 发送邮件提醒
		loginTime := time.Now().Format("2006-01-02 15:04:05")
		subject := "登录失败提醒"
		body := fmt.Sprintf(`
			<h3>登录失败提醒</h3>
			<p>您的账户正在尝试登录，但密码验证失败。</p>
			<p><b>登录邮箱:</b> %s</p>
			<p><b>登录时间:</b> %s</p>
			<p><b>登录 IP:</b> %s</p>
			<p><b>设备信息:</b> %s</p>
			<p>如果这不是您的操作，请及时修改密码。</p>
		`, email, loginTime, ip, ua)
		go utils.SendEmail(global.GVA_CONFIG.Email, []string{email}, subject, body)

		return errmsg.UserPasswordError, "", "", 0, dto.UserInfoRes{}
	}

	// 4. 登录成功，生成 Token
	jwtService := JwtService{}
	token, refreshToken, expiresAt, err := jwtService.GetTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return errmsg.CodeError, "", "", 0, dto.UserInfoRes{}
	}

	// 5. 登录成功，清除失败记录 (可选，通常建议成功后重置计数)
	global.GVA_REDIS.Del(ctx, key1m, key1h, key24h)

	// 6. 更新最后登录时间
	now := time.Now()
	model.UpdateUserLastLogin(ctx, user.ID, now)

	// 直接构造脱敏信息，减少数据库查询
	userInfo := dto.UserInfoRes{
		Username:  user.Username,
		Email:     utils.DesensitizeEmail(user.Email),
		AvatarUrl: user.AvatarUrl,
		Signature: user.Signature,
	}

	return errmsg.CodeSuccess, token, refreshToken, expiresAt, userInfo
}

// Register 用户注册
func (u *UserService) Register(ctx context.Context, username, email, password, code, signature string) (int, string, string, int64, dto.UserInfoRes) {
	// 1. 检查验证码
	keyCode := fmt.Sprintf("code:reg:%s", email)
	failKey := fmt.Sprintf("fail:reg:%s", email)

	savedCode, err := global.GVA_REDIS.Get(ctx, keyCode).Result()
	if err != nil {
		return errmsg.UserCodeNotExist, "", "", 0, dto.UserInfoRes{}
	}

	if savedCode != code {
		// 统计失败次数
		failCount, _ := global.GVA_REDIS.Incr(ctx, failKey).Result()
		if failCount == 1 {
			global.GVA_REDIS.Expire(ctx, failKey, 3*time.Minute)
		}
		if failCount > 3 {
			// 失败次数过多，清除验证码和错误记录
			global.GVA_REDIS.Del(ctx, keyCode, failKey)
			return errmsg.UserCodeLimitError, "", "", 0, dto.UserInfoRes{}
		}
		return errmsg.UserCodeError, "", "", 0, dto.UserInfoRes{}
	}

	// 2. 检查用户是否已存在
	count, _ := model.CheckUserExist(ctx, username, email)
	if count > 0 {
		return errmsg.UserAlreadyExistsError, "", "", 0, dto.UserInfoRes{}
	}

	// 3. 准备数据 (XSS 过滤与密码加密)
	safeUsername := utils.XSSFilter(username)
	safeSignature := utils.XSSFilter(signature)
	hashedPassword, _ := utils.BcryptHash(password)

	user := model.User{
		Username:     safeUsername,
		Email:        email,
		PasswordHash: hashedPassword,
		Signature:    safeSignature,
		Role:         "user", // 默认角色
		Status:       1,
	}

	// 4. 入库
	err = model.CreateUser(ctx, &user)
	if err != nil {
		return errmsg.CodeError, "", "", 0, dto.UserInfoRes{}
	}

	// 5. 注册成功，生成 Token (通过抽离出来的 JwtService)
	jwtService := JwtService{}
	token, refreshToken, expiresAt, err := jwtService.GetTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return errmsg.CodeError, "", "", 0, dto.UserInfoRes{}
	}

	// 直接构造脱敏信息，不再调用 GetUserInfo，因为数据就在内存中
	userInfo := dto.UserInfoRes{
		Username:  user.Username,
		Email:     utils.DesensitizeEmail(user.Email),
		AvatarUrl: user.AvatarUrl,
		Signature: user.Signature,
	}

	// 6. 清除验证码和错误记录
	global.GVA_REDIS.Del(ctx, keyCode, failKey)

	return errmsg.CodeSuccess, token, refreshToken, expiresAt, userInfo
}

// SendRegisterCode 发送注册验证码
func (u *UserService) SendRegisterCode(ctx context.Context, email string) int {
	// 1. 60秒限制
	key60s := fmt.Sprintf("email:reg:%s:60s", email)
	if exists, _ := global.GVA_REDIS.Exists(ctx, key60s).Result(); exists > 0 {
		return errmsg.EmailLimit60sError
	}

	// 2. 1小时限制 (max 5)
	key1h := fmt.Sprintf("email:reg:%s:1h", email)
	count1h, _ := global.GVA_REDIS.Get(ctx, key1h).Int()
	if count1h >= 5 {
		return errmsg.EmailLimit1hError
	}

	// 3. 24小时限制 (max 10)
	key24h := fmt.Sprintf("email:reg:%s:24h", email)
	count24h, _ := global.GVA_REDIS.Get(ctx, key24h).Int()
	if count24h >= 10 {
		return errmsg.EmailLimit24hError
	}

	// 生成 4 位验证码
	code := utils.GenerateRandomCode(4)

	// 发送邮件
	subject := "注册验证码"
	body := fmt.Sprintf("您的注册验证码为: <b>%s</b>，有效期为 3 分钟，请勿泄露给他人。", code)
	err := utils.SendEmail(global.GVA_CONFIG.Email, []string{email}, subject, body)
	if err != nil {
		fmt.Printf("发送邮件失败: %v\n", err)
		return errmsg.EmailSendFailedError
	}

	// 记录频率
	// 60秒
	global.GVA_REDIS.Set(ctx, key60s, "1", 60*time.Second)

	// 1小时
	global.GVA_REDIS.Incr(ctx, key1h)
	if count1h == 0 {
		global.GVA_REDIS.Expire(ctx, key1h, time.Hour)
	}

	// 24小时
	global.GVA_REDIS.Incr(ctx, key24h)
	if count24h == 0 {
		global.GVA_REDIS.Expire(ctx, key24h, 24*time.Hour)
	}

	// 存储验证码 (注册场景, 3分钟有效期)
	keyCode := fmt.Sprintf("code:reg:%s", email)
	global.GVA_REDIS.Set(ctx, keyCode, code, 3*time.Minute)

	return errmsg.CodeSuccess
}
