package controllers

import (
	"exchangeapp/config"
	"exchangeapp/global"
	"exchangeapp/models"
	"exchangeapp/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func Register(ctx *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"` // 可选
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"email"`
		Avatar   string `json:"avatar"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色合法性
	if input.Role != "" && input.Role != "admin" && input.Role != "user" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role value"})
		return
	}

	if !utils.ValidatePassword(input.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "密码至少8位，必须包含大写字母、小写字母和数字",
		})
		return
	}

	hashedPwd, err := utils.HashPassword(input.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username: input.Username,
		Password: hashedPwd,
		Role:     input.Role,
		Nickname: input.Nickname,
		Email:    input.Email,
		Avatar:   input.Avatar,
	}

	if user.Role == "" {
		user.Role = "user"
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if err := global.Db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
		if existingUser.Username == user.Username {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已存在"})
		}
		return
	}

	if err := global.Db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(user.Username, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = "" // 不返回密码
	ctx.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func Login(ctx *gin.Context) {
	// 登录过程的用户名和密码
	var input struct {
		Username string `json:"username"` // 结构体标签(方便将json字段映射到结构体)
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User

	// 验证用户名
	if err := global.Db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "wrong credentials",
		})
		return
	}

	key := "login_fail:" + input.Username
	failCount, _ := global.RedisDB.Get(ctx, key).Int()
	if failCount >= 5 {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "尝试次数过多，请15分钟后再试"})
		return
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		// 登录失败时
		global.RedisDB.Incr(ctx, key)
		global.RedisDB.Expire(ctx, key, time.Minute*15)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "wrong credentials",
		})
		return
	}

	// 登录成功，删除失败计数
	global.RedisDB.Del(ctx, key)

	token, err := utils.GenerateJWT(user.Username, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// 新增获取/更新用户信息接口
func GetProfile(ctx *gin.Context) {
	username := ctx.GetString("username")
	var user models.User
	if err := global.Db.Where("username = ?", username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.Password = "" // 不返回密码
	ctx.JSON(http.StatusOK, user)
}

func UpdateProfile(ctx *gin.Context) {
	username := ctx.GetString("username")

	var input struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"omitempty,email"`
		Avatar   string `json:"avatar"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := global.Db.Where("username = ?", username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if input.Nickname != "" {
		user.Nickname = input.Nickname
	}

	if input.Email != "" {
		user.Email = input.Email
	}

	if input.Avatar != "" {
		user.Avatar = input.Avatar
	}

	if input.Password != "" {
		// 校验密码强度
		if !utils.ValidatePassword(input.Password) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "密码至少8位，必须包含大写字母、小写字母和数字",
			})
			return
		}

		hashedPwd, err := utils.HashPassword(input.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user.Password = hashedPwd
	}

	if err := global.Db.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusOK, user)
}

// 上传用户头像
func UploadAvatar(ctx *gin.Context) {
	// 从上下文中获取当前用户 ID（假设你在中间件里已经设置了 userID）
	userID := ctx.GetUint("userID")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}

	// 限制文件类型（可选）
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
		return
	}

	// 创建保存路径，如果不存在则创建
	saveDir := "./uploads/avatars"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
			return
		}
	}

	// 文件命名：用户ID + 时间戳 + 后缀，防止重名
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	savePath := filepath.Join(saveDir, filename)

	// 保存文件到本地
	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 更新数据库中的 Avatar 字段
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 从配置中获取基础URL
	baseURL := fmt.Sprintf("http://localhost:%s", config.AppConfig.App.Port)
	avatarURL := baseURL + "/uploads/avatars/" + filename

	user.Avatar = avatarURL
	if err := global.Db.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户头像失败"})
		return
	}

	// 返回成功和新头像路径
	ctx.JSON(http.StatusOK, gin.H{
		"message": "头像上传成功",
		"avatar":  avatarURL, // 返回完整URL
	})
}
