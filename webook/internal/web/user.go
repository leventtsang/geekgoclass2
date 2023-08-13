package web

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"time"
)

// UserHandler 我准备在它上面定义跟用户有关的路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	userService *service.UserService
}

type UserEditRequest struct {
	Nickname string `json:"nickname" binding:"required,max=100"`
	Birthday string `json:"birthday" binding:"required,datetime=2006-01-02"`
	Bio      string `json:"bio" binding:"max=500"`
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		userService: svc,
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

//func (u *UserHandler) RegisterRoutesV1(ug *gin.RouterGroup) {
//	ug.GET("/profile", u.Profile)
//	ug.POST("/signup", u.SignUp)
//	ug.POST("/login", u.Login)
//	ug.POST("/edit", u.Edit)
//}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.EditProfile)
	ug.POST("/login", u.LoginJWT)
	//ug.POST("/edit", u.Edit)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// 先绑定请求数据
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusBadRequest, "请求格式错误")
		return
	}

	// 现在进行邮箱和密码的验证
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusBadRequest, "你的邮箱格式不对")
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusBadRequest, "两次输入的密码不一致")
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusBadRequest, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 现在我们可以尝试注册用户
	user := domain.User{
		Email:    req.Email,
		Password: req.Password,
	}

	log.Printf("Before check: user.Birthday = %v", user.Birthday)

	// 检查并处理Birthday字段
	if user.Birthday != nil && user.Birthday.IsZero() {
		user.Birthday = nil
	}
	log.Printf("After check: user.Birthday = %v", user.Birthday)

	err = u.svc.SignUp(ctx, user)
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里用 JWT 设置登录态
	// 生成一个 JWT token

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		// 一分钟过期
		MaxAge: 60,
	})
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) GetUserProfileByID(userID int64) (*domain.User, error) {
	fmt.Printf("Getting user profile for userID: %d\n", userID) // 添加分析日志
	// 在这里调用 userService 的 GetUserProfileByID 方法
	user, err := u.userService.GetUserProfileByID(userID) // 确保方法名正确
	if err != nil {
		fmt.Printf("Error fetching user profile: %v\n", err) // 添加分析日志
		return nil, err
	}
	fmt.Printf("Retrieved user profile: %v\n", user) // 添加分析日志
	return user, nil
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, exists := ctx.Get("claims")
	if !exists {
		ctx.String(http.StatusInternalServerError, "无法获取用户信息")
		return
	}

	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "无效的用户信息")
		return
	}

	// 调用服务层方法获取用户资料
	user, err := u.userService.GetUserProfileByID(claims.Uid)
	if err != nil {
		log.Printf("Error fetching user profile: %v", err)
		ctx.String(http.StatusInternalServerError, "获取用户资料失败")
		return
	}

	log.Printf("User Profile: %+v", user)

	// 返回用户资料
	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	// 获取用户ID，根据您的应用逻辑实现 getCurrentUserID 函数
	userID := getCurrentUserID(ctx)

	// 使用 userService 获取用户信息
	user, err := u.userService.GetUserByID(userID)
	if err != nil {
		// 处理错误
		ctx.String(http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// 返回用户信息
	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) EditProfile(ctx *gin.Context) {
	var req UserEditRequest

	// 尝试绑定JSON数据
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON provided: " + err.Error()})
		return
	}

	userID := getCurrentUserID(ctx)
	log.Printf("Editing profile for userID: %d", userID)

	if err := h.userService.UpdateUserProfile(userID, req.Nickname, req.Birthday, req.Bio); err != nil {
		log.Printf("Error updating user profile: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	log.Printf("Profile updated successfully for userID: %d", userID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *UserHandler) GetProfile(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func getCurrentUserID(ctx *gin.Context) int64 {
	claims, exists := ctx.Get("claims")
	if !exists {
		// 无法获取用户信息，返回合适的错误或默认值
		return 0
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		// 无效的用户信息，返回合适的错误或默认值
		return 0
	}

	// 返回解析出来的用户ID
	return userClaims.Uid
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明你自己的要放进去 token 里面的数据
	Uid int64
	// 自己随便加
	UserAgent string
}
