package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取所有用户列表 (仅管理员)
func GetUserList(ctx *gin.Context) {
	var users []models.User
	// 查询所有用户，排除密码字段
	// Select 语法取决于 GORM 版本，这里使用简单查询后手动清空密码，或者使用 Smart Select
	if err := global.Db.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 处于安全考虑，清空密码和敏感信息
	var safeUsers []map[string]interface{}
	for _, u := range users {
		safeUsers = append(safeUsers, map[string]interface{}{
			"ID":        u.ID,
			"username":  u.Username,
			"nickname":  u.Nickname,
			"email":     u.Email,
			"role":      u.Role,
			"avatar":    u.Avatar,
			"CreatedAt": u.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, safeUsers)
}

// UpdateUserRole 修改用户角色 (仅管理员)
func UpdateUserRole(ctx *gin.Context) {
	// 1. 获取目标用户 ID
	targetUserID := ctx.Param("id")

	// 2. 获取请求体中的新角色
	var input struct {
		Role string `json:"role" binding:"required"` // 只允许 "admin" 或 "user"
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 简单校验角色合法性
	if input.Role != "admin" && input.Role != "user" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role value"})
		return
	}

	// 3. 查找并更新用户
	var user models.User
	if err := global.Db.First(&user, targetUserID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 防止自我降级 (可选：防止管理员不小心把自己改成 user 后失去权限)
	currentUsername := ctx.GetString("username")
	if user.Username == currentUsername && input.Role == "user" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You cannot demote yourself"})
		return
	}

	// 更新数据库
	if err := global.Db.Model(&user).Update("role", input.Role).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Role updated successfully",
		"user_id":  user.ID,
		"new_role": input.Role,
	})
}

// DeleteUser 删除用户 (仅管理员)
func DeleteUser(ctx *gin.Context) {
	// 目标用户 ID（要删除谁）
	targetID := ctx.Param("id")

	// 当前正在操作的管理员 ID（从 Token 获取）
	currentUserID := ctx.GetString("userID")

	// 防止删除自己
	if currentUserID == targetID {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "不能删除自己的账号",
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := global.Db.First(&user, targetID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 删除用户
	if err := global.Db.Delete(&models.User{}, targetID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
