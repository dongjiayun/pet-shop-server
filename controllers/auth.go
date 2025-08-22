package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dongjiayun/pet-shop-server/config"
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/dongjiayun/pet-shop-server/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/google/uuid"
)

type MyClaims struct {
	Cid       string `json:"cid"`
	LoginType string `json:"login_type"`
	jwt.StandardClaims
}

const TokenExpireDuration = config.TokenExpireDuration

var Secret = []byte(config.Secret)

func SignIn(c *gin.Context) {
	var user models.AuthUser
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	if user.LoginType == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "请选择登录方式"})
	}
	signup := func() {
		// 邮箱不存在时，先发送验证码进行注册
		if user.Otp == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
			return
		}

		if user.Ticket == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "ticket不能为空"})
			return
		}

		// 验证验证码是否正确
		optCache := models.RedisClient.Get(context.Background(), "signup"+user.Email)
		if optCache.Val() != "" {
			var cache models.AuthOtp
			json.Unmarshal([]byte(optCache.Val()), &cache)
			if cache.Code == user.Otp && cache.Ticket == user.Ticket {
				// 验证码正确，创建用户
				ch := make(chan string)
				go CreateByEmail(ch, c, user.Email)
				result := <-ch
				if result == "success" {
					generateToken(c, user.Email, "email")
					models.RedisClient.Del(context.Background(), "signup"+user.Email)
					return
				} else {
					c.JSON(200, models.Result{Code: 10001, Message: result})
					return
				}
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "验证码错误"})
				return
			}
		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "请先发送验证码"})
			return
		}
	}
	switch user.LoginType {
	case "signup":
		signup()
	case "phone":
		if user.Phone == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "手机号不能为空"})
			return
		}
		if user.Otp == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
		}
		phoneExist := checkPhoneExists(user.Phone, "")
		if phoneExist {

		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "手机号不存在"})
		}
	case "email":
		if user.Email == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
			return
		}
		emailExist := checkEmailExists(user.Email, "")
		if emailExist {
			if user.Otp == "" {
				c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
				return
			}

			if user.Ticket == "" {
				c.JSON(200, models.Result{Code: 10001, Message: "ticket不能为空"})
				return
			}

			optCache := models.RedisClient.Get(context.Background(), user.Email)

			if optCache.Val() != "" {
				var cache models.AuthOtp
				json.Unmarshal([]byte(optCache.Val()), &cache)
				if cache.Code == user.Otp && cache.Ticket == user.Ticket {
					generateToken(c, user.Email, "email")
					models.RedisClient.Del(context.Background(), user.Email)
					return
				} else {
					c.JSON(200, models.Result{Code: 10001, Message: "验证码错误"})
					return
				}
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "请发送验证码"})
			}
		} else {
			signup()
		}
	case "emailWithPassword":
		if user.Email == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
			return
		}
		if user.Password == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "密码不能为空"})
			return
		}
		emailExist := checkEmailExists(user.Email, "")
		if emailExist {
			var resultUser models.User
			db := models.DB.Model(&models.User{}).Where("email = ?", user.Email).First(&resultUser)
			if db.Error != nil {
				// SQL执行失败，返回错误信息
				c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
				return
			}
			if resultUser.Password == user.Password {
				if user.Password == "123456" {
					c.JSON(200, models.Result{Code: 10001, Message: "初始密码无法用于登陆,请您使用邮箱验证码登录后修改后重试~"})
					return
				}
				generateToken(c, user.Email, "email")
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "密码错误"})
			}
		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		}
	case "wechat":
		client := &http.Client{}

		appid := config.MiniprogramAppid

		secret := config.MiniprogramSecret

		js_code := user.JsCode

		params := url.Values{}
		params.Set("appid", appid)
		params.Set("secret", secret)
		params.Set("js_code", js_code)
		params.Set("grant_type", "authorization_code")

		queryString := params.Encode()

		req, err := http.NewRequest("GET", "https://api.weixin.qq.com/sns/jscode2session?"+queryString, nil)

		req.Header.Add("Content-Type", "application/json")

		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(200, models.Result{Code: 10001, Message: "internal server error"})
			return
		}

		type Resp struct {
			Openid     string `json:"openid"`
			SessionKey string `json:"session_key"`
			Unionid    string `json:"unionid"`
		}

		var data Resp

		err = json.Unmarshal(body, &data)

		openId := data.Openid
		unionId := data.Unionid

		openidExists := checkOpenidExists(openId)

		if openidExists {
			generateToken(c, openId, "wechat")
		} else {
			ch := make(chan string)
			go CreateByOpenid(ch, c, openId, unionId)
			result := <-ch
			if result == "success" {
				generateToken(c, openId, "wechat")
			}
		}
	}
}

func generateToken(c *gin.Context, account string, loginType string) {
	var resultUser models.User

	if loginType == "email" {
		db := models.DB.Model(&resultUser).Where("email = ?", account).First(&resultUser)
		if db.Error != nil {
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	} else if loginType == "wechat" {
		db := models.DB.Model(&resultUser).Where("openid = ?", account).First(&resultUser)
		if db.Error != nil {
			// SQL执行失败，返回错误信息
			c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		}
	}

	token, _ := GenToken(resultUser.Cid, loginType)

	refreshToken, _ := GenRefreshToken(resultUser.Cid, loginType)

	type Result struct {
		models.SafeUser
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}

	result := Result{
		SafeUser:     models.GetSafeUser(resultUser),
		Token:        token,
		RefreshToken: refreshToken,
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: result})
}

func SendEmailOtp(c *gin.Context) {
	type OtpCode struct {
		Email string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	}
	var otpCode OtpCode
	err := c.ShouldBindJSON(&otpCode)
	if otpCode.Email != "" && err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &otpCode)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if otpCode.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	emailExist := checkEmailExists(otpCode.Email, "")
	if !emailExist {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		return
	}

	optCache := models.RedisClient.Get(context.Background(), otpCode.Email)

	if optCache.Val() != "" {
		var cache models.AuthOtp
		json.Unmarshal([]byte(optCache.Val()), &cache)
		c.JSON(200, models.Result{
			Code:    10001,
			Message: "验证码已发送,请勿重复发送",
			Data:    cache.Ticket,
		})
		return
	}

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(900000) + 100000
	randomNumberStr := strconv.Itoa(randomNumber)

	to := otpCode.Email
	subject := "BIRKIN PET 邮箱登陆验证码"
	message := "尊敬的用户：\n\n您好！您正在进行的操作需要验证身份。\n验证码：" + randomNumberStr + "\n（一分钟之内有效）\n\n请勿向他人泄露此验证码。"

	smtpHost := config.SmtpHost
	smtpPort := config.SmtpPort
	smtpUser := config.SmtpUser
	smtpPassword := config.SmtpPassword

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)   // 发件人邮箱
	m.SetHeader("To", to)           // 收件人邮箱
	m.SetHeader("Subject", subject) // 邮件主题
	m.SetBody("text/html", message) // 邮件内容

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ticket := uuid.New()
	ticketString := ticket.String()

	authOtp := models.AuthOtp{
		Code:    randomNumberStr,
		Account: otpCode.Email,
		Ticket:  ticketString,
	}

	authOtpJSON, _ := json.Marshal(authOtp)

	redisClient := models.RedisClient

	msg := redisClient.Set(context.Background(), otpCode.Email, authOtpJSON, 1*time.Minute)

	if msg != nil {
		fmt.Println(msg)
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: ticketString})
}

func SendSignupEmailOtp(c *gin.Context) {
	type OtpCode struct {
		Email string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	}
	var otpCode OtpCode
	err := c.ShouldBindJSON(&otpCode)
	if otpCode.Email != "" && err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &otpCode)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if otpCode.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	emailExist := checkEmailExists(otpCode.Email, "")
	if emailExist {
		c.JSON(200, models.Result{Code: 10001, Message: "该邮箱已被注册"})
		return
	}

	optCache := models.RedisClient.Get(context.Background(), "signup"+otpCode.Email)

	if optCache.Val() != "" {
		var cache models.AuthOtp
		json.Unmarshal([]byte(optCache.Val()), &cache)
		c.JSON(200, models.Result{
			Code:    10001,
			Message: "验证码已发送,请勿重复发送",
			Data:    cache.Ticket,
		})
		return
	}

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(900000) + 100000
	randomNumberStr := strconv.Itoa(randomNumber)

	to := otpCode.Email
	subject := "BIRKIN PET 邮箱注册验证码"
	message := "尊敬的用户：\n\n您好！您正在进行的操作需要验证身份。\n验证码：" + randomNumberStr + "\n（一分钟之内有效）\n\n请勿向他人泄露此验证码。"

	smtpHost := config.SmtpHost
	smtpPort := config.SmtpPort
	smtpUser := config.SmtpUser
	smtpPassword := config.SmtpPassword

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)   // 发件人邮箱
	m.SetHeader("To", to)           // 收件人邮箱
	m.SetHeader("Subject", subject) // 邮件主题
	m.SetBody("text/html", message) // 邮件内容

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ticket := uuid.New()
	ticketString := ticket.String()

	authOtp := models.AuthOtp{
		Code:    randomNumberStr,
		Account: otpCode.Email,
		Ticket:  ticketString,
	}

	authOtpJSON, _ := json.Marshal(authOtp)

	redisClient := models.RedisClient

	msg := redisClient.Set(context.Background(), "signup"+otpCode.Email, authOtpJSON, 1*time.Minute)

	if msg != nil {
		fmt.Println(msg)
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: ticketString})
}

func SendResetEmailOtp(c *gin.Context) {
	type OtpCode struct {
		Email string `json:"email" binding:"email" msg:"请输入正确的邮箱地址" gorm:"index"`
	}
	var otpCode OtpCode
	err := c.ShouldBindJSON(&otpCode)
	if otpCode.Email != "" && err != nil {
		// 显示自定义的错误信息
		msg := utils.GetValidMsg(err, &otpCode)
		c.JSON(200, models.Result{Code: 10001, Message: msg})
		return
	}
	if otpCode.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	emailExist := checkEmailExists(otpCode.Email, "")
	if !emailExist {
		c.JSON(200, models.Result{Code: 10001, Message: "该邮箱不存在"})
		return
	}

	optCache := models.RedisClient.Get(context.Background(), "signup"+otpCode.Email)

	if optCache.Val() != "" {
		var cache models.AuthOtp
		json.Unmarshal([]byte(optCache.Val()), &cache)
		c.JSON(200, models.Result{
			Code:    10001,
			Message: "验证码已发送,请勿重复发送",
			Data:    cache.Ticket,
		})
		return
	}

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(900000) + 100000
	randomNumberStr := strconv.Itoa(randomNumber)

	to := otpCode.Email
	subject := "BIRKIN PET 重置邮箱账号密码"
	message := "尊敬的用户：\n\n您好！您正在进行的操作需要验证身份。\n验证码：" + randomNumberStr + "\n（一分钟之内有效）\n\n请勿向他人泄露此验证码。"

	smtpHost := config.SmtpHost
	smtpPort := config.SmtpPort
	smtpUser := config.SmtpUser
	smtpPassword := config.SmtpPassword

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)   // 发件人邮箱
	m.SetHeader("To", to)           // 收件人邮箱
	m.SetHeader("Subject", subject) // 邮件主题
	m.SetBody("text/html", message) // 邮件内容

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ticket := uuid.New()
	ticketString := ticket.String()

	authOtp := models.AuthOtp{
		Code:    randomNumberStr,
		Account: otpCode.Email,
		Ticket:  ticketString,
	}

	authOtpJSON, _ := json.Marshal(authOtp)

	redisClient := models.RedisClient

	msg := redisClient.Set(context.Background(), "resetPassword"+otpCode.Email, authOtpJSON, 1*time.Minute)

	if msg != nil {
		fmt.Println(msg)
	}

	c.JSON(200, models.Result{Code: 0, Message: "success", Data: ticketString})
}

type WastedToken struct {
	Token      string `json:"token"`
	CreateTime int    `json:"create_time"`
}

func SignOut(c *gin.Context) {
	token := c.GetHeader("Authorization")
	handleWasteToken(token)
	c.JSON(200, models.Result{Code: 0, Message: "success"})
}

func handleWasteToken(token string) {
	redisClient := models.RedisClient
	blackList := redisClient.Get(context.Background(), "blackList")
	blackListValue := blackList.Val()
	var _blackList []WastedToken
	if blackListValue != "" {
		redisClient.Set(context.Background(), "blackList", "", 0)
	}
	err := json.Unmarshal([]byte(blackListValue), &_blackList)
	if err != nil {
		// 处理解析错误
		fmt.Println("解析JSON出错:", err)
		// 返回错误或者其他逻辑处理
	}
	wastedToken := WastedToken{
		Token:      token,
		CreateTime: int(time.Now().Unix()),
	}
	_blackList = append(_blackList, wastedToken)
	__blackList, _ := json.Marshal(_blackList)
	redisClient.Set(context.Background(), "blackList", __blackList, 0)
}

func RefreshToken(c *gin.Context) {
	refreshToken, _ := CheckRefreshToken(c)
	if refreshToken == nil {
		c.JSON(200, models.Result{Code: 10001, Message: "无效的RefreshToken"})
		return
	}
	cid := refreshToken.Cid
	fmt.Println(cid)
	loginType := refreshToken.LoginType
	var user models.User
	db := models.DB.Model(&models.User{}).Where("cid = ?", cid).First(&user)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	generateToken(c, user.Email, loginType)
}

func CheckRefreshToken(c *gin.Context) (*TokenClaims, error) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		return nil, errors.New("无效的RefreshToken")
	}

	_, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(refreshToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyClaims)

	cid := claims.Cid
	loginType := claims.LoginType

	fmt.Println(cid, loginType)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &TokenClaims{
		Cid:       cid,
		LoginType: loginType,
	}, nil
}

// GenToken 生成JWT
func GenToken(Cid string, LoginType string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		Cid, // 自定义字段
		LoginType,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "pet-family",                               // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tokenString, err := token.SignedString(Secret)
	if err != nil {
		return "", err
	}

	// 存储token和cid的映射关系到Redis
	redisClient := models.RedisClient
	// 使用token作为key，cid作为value，并设置与token相同的过期时间
	redisClient.Set(context.Background(), tokenString, Cid, TokenExpireDuration)

	return tokenString, nil
}

func GenRefreshToken(Cid string, LoginType string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		Cid, // 自定义字段
		LoginType,
		jwt.StandardClaims{
			Issuer: "pet-family", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(Secret)
}

type TokenClaims struct {
	Cid       string
	LoginType string
}

func CheckToken(c *gin.Context) (*TokenClaims, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return nil, errors.New("无效的token")
	}

	redisClient := models.RedisClient
	blackList := redisClient.Get(context.Background(), "blackList")
	blackListValue := blackList.Val()
	var _blackList []WastedToken
	if blackListValue != "" {
		err := json.Unmarshal([]byte(blackListValue), &_blackList)
		if err != nil {
			// 处理解析错误
			fmt.Println("解析JSON出错:", err)
			// 返回错误或者其他逻辑处理
		}
		var isWasted bool
		var newBlackList []WastedToken
		for _, wasted := range _blackList {
			if wasted.Token == tokenString {
				isWasted = true
			}
			nowTime := int(time.Now().Unix())
			expiredTime := int(time.Unix(int64(wasted.CreateTime), 0).Add(TokenExpireDuration).Unix())
			if nowTime < expiredTime {
				newBlackList = append(newBlackList, wasted)
			}
		}
		__blackList, _ := json.Marshal(newBlackList)
		redisClient.Set(context.Background(), "blackList", __blackList, 0)
		if isWasted {
			return nil, errors.New("token已失效")
		}
	}

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 获取用户的CID（假设你的 MyClaims 结构体中有一个 CID 字段）
	cid := claims.Cid

	// 返回 TokenClaims 结构体，包含 MyClaims 和 CID
	return &TokenClaims{
		Cid: cid,
	}, nil
}

func CheckSelfOrAdmin(c *gin.Context, cid string, ch chan string) {
	claims, err := CheckToken(c)
	var message string
	if err != nil {
		message = "请重新登陆"
	}
	fmt.Println(claims.Cid)
	if claims.Cid == "C000000000001" {
		message = "success"
	} else if claims.Cid == cid {
		message = "success"
	} else {
		message = "您没有该操作的权限"
	}

	ch <- message
}

type ResetPasswordRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

func ResetPassword(c *gin.Context) {
	cid, _ := c.Get("cid")
	var request ResetPasswordRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	password := request.Password
	newPassword := request.NewPassword

	if !utils.CheckPassword(newPassword) {
		c.JSON(200, models.Result{Code: 10002, Message: "请输入6-20位密码,必须包含数字和字母"})
		return
	}

	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)

	if password != user.Password {
		c.JSON(200, models.Result{Code: 10003, Message: "原密码不正确"})
		return
	}

	models.DB.Model(&user).Where("cid = ?", cid).Update("password", newPassword)

	token := c.GetHeader("Authorization")
	handleWasteToken(token)

	c.JSON(200, models.Result{0, "success", nil})
}

type FindbackPasswordRequest struct {
	Password string `json:"password"`
	Smscode  string `json:"smscode"`
	Email    string `json:"email"`
	Ticket   string `json:"ticket"`
}

func FindbackPassword(c *gin.Context) {
	var request FindbackPasswordRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	if request.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}
	emailExist := checkEmailExists(request.Email, "")
	if emailExist {
		if request.Smscode == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
			return
		}

		if request.Ticket == "" {
			c.JSON(200, models.Result{Code: 10001, Message: "ticket不能为空"})
			return
		}

		optCache := models.RedisClient.Get(context.Background(), request.Email)

		if optCache.Val() != "" {
			var cache models.AuthOtp
			json.Unmarshal([]byte(optCache.Val()), &cache)
			if cache.Code == request.Smscode && cache.Ticket == request.Ticket {
				var user models.User
				models.DB.Where("email = ?", request.Email).First(&user)
				user.Password = request.Password
				models.DB.Save(&user)
				token := c.GetHeader("Authorization")
				if token != "" {
					handleWasteToken(token)
				}
				c.JSON(200, models.Result{0, "success", nil})
				return
			} else {
				c.JSON(200, models.Result{Code: 10001, Message: "验证码错误"})
				return
			}
		} else {
			c.JSON(200, models.Result{Code: 10001, Message: "请发送验证码"})
		}
	} else {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		return
	}
}

// GetCidByToken 通过token获取用户CID
func GetCidByToken(c *gin.Context) (string, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return "", errors.New("无效的token")
	}

	// 先从Redis中查询token对应的cid
	redisClient := models.RedisClient
	cid, err := redisClient.Get(context.Background(), tokenString).Result()
	if err == nil && cid != "" {
		// Redis中存在该token的映射关系，直接返回cid
		return cid, nil
	}

	// 检查token是否在黑名单中
	blackList := redisClient.Get(context.Background(), "blackList")
	blackListValue := blackList.Val()
	var _blackList []WastedToken
	if blackListValue != "" {
		err := json.Unmarshal([]byte(blackListValue), &_blackList)
		if err != nil {
			return "", err
		}

		// 检查token是否在黑名单中
		for _, wasted := range _blackList {
			if wasted.Token == tokenString {
				return "", errors.New("token已失效")
			}
		}
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Secret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		// 将token和cid的映射关系存入Redis，设置与token相同的过期时间
		expireTime := time.Duration(claims.ExpiresAt-time.Now().Unix()) * time.Second
		redisClient.Set(context.Background(), tokenString, claims.Cid, expireTime)
		return claims.Cid, nil
	}
	return "", errors.New("无效的token")
}

func CheckIsAdmin(c *gin.Context) bool {
	cid, _ := c.Get("cid")
	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)
	return user.Role == 2 || user.Role == 3
}

func CheckIsSuperAdmin(c *gin.Context) bool {
	cid, _ := c.Get("cid")
	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)
	return user.Role == 3
}

func CheckIsEmployee(c *gin.Context) bool {
	cid, _ := c.Get("cid")
	var user models.User
	models.DB.Where("cid = ?", cid).First(&user)
	return !(user.Role == 0 || user.Role == 4)
}

type ResetPasswordByOtpRequest struct {
	Email    string `json:"email"`
	Otp      string `json:"otp"`
	Ticket   string `json:"ticket"`
	Password string `json:"password"`
}

func ResetPasswordByOtp(c *gin.Context) {
	var request ResetPasswordByOtpRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}

	if request.Email == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不能为空"})
		return
	}

	if request.Otp == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "验证码不能为空"})
		return
	}

	if request.Ticket == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "ticket不能为空"})
		return
	}

	if request.Password == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "新密码不能为空"})
		return
	}

	// 验证密码格式
	if !utils.CheckPassword(request.Password) {
		c.JSON(200, models.Result{Code: 10002, Message: "请输入6-20位密码,必须包含数字和字母"})
		return
	}

	// 检查邮箱是否存在
	emailExist := checkEmailExists(request.Email, "")
	if !emailExist {
		c.JSON(200, models.Result{Code: 10001, Message: "邮箱不存在"})
		return
	}

	// 验证验证码
	optCache := models.RedisClient.Get(context.Background(), "resetPassword"+request.Email)
	if optCache.Val() == "" {
		c.JSON(200, models.Result{Code: 10001, Message: "请先发送验证码"})
		return
	}

	var cache models.AuthOtp
	err = json.Unmarshal([]byte(optCache.Val()), &cache)
	if err != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	if cache.Code != request.Otp || cache.Ticket != request.Ticket {
		c.JSON(200, models.Result{Code: 10001, Message: "验证码错误"})
		return
	}

	// 更新密码
	var user models.User
	db := models.DB.Model(&models.User{}).Where("email = ?", request.Email).First(&user)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	// 更新密码
	user.Password = request.Password
	db = models.DB.Save(&user)
	if db.Error != nil {
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	// 删除已使用的验证码
	models.RedisClient.Del(context.Background(), "resetPassword"+request.Email)

	// 使当前token失效（如果用户已登录）
	token := c.GetHeader("Authorization")
	if token != "" {
		handleWasteToken(token)
	}

	c.JSON(200, models.Result{Code: 0, Message: "密码重置成功"})
}
