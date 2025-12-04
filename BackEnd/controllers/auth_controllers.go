package controllers

import (
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

	// 将JSON内容映射到input结构体
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
	failCount, _ := global.RedisDB.Get(ctx, key).Int() // 执行 Redis GET 命令获取指定 key 的值，将获取到的字符串值转换为整数
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

func UploadAvatar(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	// 简单的扩展名检查
	ext := filepath.Ext(file.Filename)
	allowExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "只允许上传 jpg, png, gif 图片"})
		return
	}

	// 确保存储目录存在
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 返回相对路径 URL
	url := "/uploads/avatars/" + filename
	ctx.JSON(http.StatusOK, gin.H{"url": url})
}
