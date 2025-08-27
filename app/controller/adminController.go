package controller

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/middleware"
	"gameWeb/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("管理员登录参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 记录登录IP
	clientIP := c.ClientIP()

	// 1. 查询管理员账户
	admin, err := getAdminByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnf("管理员登录失败 - 账户不存在: %s, IP: %s", req.Username, clientIP)
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Code:    401,
				Message: "用户名或密码错误",
			})
			return
		}
		log.Errorf("查询管理员账户失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 2. 验证账户状态
	if admin.Status != 1 {
		log.Warnf("管理员账户已禁用: %s, IP: %s", req.Username, clientIP)
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "账户已被禁用",
		})
		return
	}

	// 3. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		log.Warnf("管理员登录失败 - 密码错误: %s, IP: %s", req.Username, clientIP)
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	// 4. 更新最后登录信息
	admin.LastLoginIP = clientIP
	admin.LastLoginTime = time.Now()
	if err := updateAdminLoginInfo(admin); err != nil {
		log.Errorf("更新管理员登录信息失败: %v", err)
	}

	// 5. 生成JWT Token
	token, err := middleware.GenerateAdminJWT(admin)
	if err != nil {
		log.Errorf("生成管理员JWT失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 6. 缓存会话到Redis
	sessionKey := fmt.Sprintf("admin_session:%d", admin.ID)
	sessionData := fmt.Sprintf("active:%d", time.Now().Unix())
	expiration := time.Duration(config.AppConfig.Admin.SessionTimeout) * time.Hour

	if err := db.SetRedisWithExpire(sessionKey, sessionData, expiration); err != nil {
		log.Errorf("缓存管理员会话失败: %v", err)
		// 不返回错误，登录仍然可以继续，但会话验证可能会失败
	}

	// 7. 构建响应
	adminInfo := &models.AdminInfo{
		ID:            admin.ID,
		Username:      admin.Username,
		Email:         admin.Email,
		RealName:      admin.RealName,
		IsSuperAdmin:  admin.IsSuperAdmin == 1,
		LastLoginTime: admin.LastLoginTime,
	}

	response := models.AdminLoginResponse{
		Token:     token,
		AdminInfo: adminInfo,
	}

	log.Infof("管理员登录成功: %s (ID: %d), IP: %s", admin.Username, admin.ID, clientIP)
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "登录成功",
		Data:    response,
	})
}

// AdminLogout 管理员登出
func AdminLogout(c *gin.Context) {
	// 从上下文获取管理员ID
	adminId, exists := c.Get("adminId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	// 删除Redis中的会话
	sessionKey := fmt.Sprintf("admin_session:%d", adminId)
	if err := db.DelRedis(sessionKey); err != nil {
		log.Errorf("删除管理员会话失败: %v", err)
	}

	log.Infof("管理员登出成功: ID: %v, IP: %s", adminId, c.ClientIP())
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "登出成功",
	})
}

// GetAdminInfo 获取当前管理员信息
func GetAdminInfo(c *gin.Context) {
	// 从上下文获取管理员信息
	adminId, exists := c.Get("adminId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	// 查询完整的管理员信息
	admin, err := getAdminByID(adminId.(uint64))
	if err != nil {
		log.Errorf("查询管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	adminInfo := &models.AdminInfo{
		ID:            admin.ID,
		Username:      admin.Username,
		Email:         admin.Email,
		RealName:      admin.RealName,
		IsSuperAdmin:  admin.IsSuperAdmin == 1,
		LastLoginTime: admin.LastLoginTime,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    adminInfo,
	})
}

// CreateAdmin 创建管理员账户（仅超级管理员可用）
func CreateAdmin(c *gin.Context) {
	var req struct {
		Username     string `json:"username" binding:"required,min=3,max=50"`
		Password     string `json:"password" binding:"required,min=6,max=50"`
		Email        string `json:"email" binding:"required,email"`
		RealName     string `json:"realName" binding:"required,min=1,max=50"`
		Mobile       string `json:"mobile"`
		IsSuperAdmin int8   `json:"isSuperAdmin"`
		DepartmentID *int   `json:"departmentId"`
		Note         string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("创建管理员参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	if exists, err := checkAdminUsernameExists(req.Username); err != nil {
		log.Errorf("检查用户名失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	} else if exists {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "用户名已存在",
		})
		return
	}

	// 检查邮箱是否已存在
	if exists, err := checkAdminEmailExists(req.Email); err != nil {
		log.Errorf("检查邮箱失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	} else if exists {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "邮箱已存在",
		})
		return
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("生成密码哈希失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 获取创建者ID
	createdBy, _ := c.Get("adminId")
	createdByID := createdBy.(uint64)

	// 创建管理员
	admin := &models.AdminAccount{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Email:        req.Email,
		Mobile:       req.Mobile,
		Status:       1,
		IsSuperAdmin: req.IsSuperAdmin,
		RealName:     req.RealName,
		DepartmentID: req.DepartmentID,
		Note:         req.Note,
		CreatedBy:    &createdByID,
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}

	if err := createAdmin(admin); err != nil {
		log.Errorf("创建管理员失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "创建失败",
		})
		return
	}

	log.Infof("管理员账户创建成功: %s, 创建者: %d", req.Username, createdByID)
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "创建成功",
		Data: gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
	})
}

// 数据库操作函数

// getAdminByUsername 根据用户名查询管理员
func getAdminByUsername(username string) (*models.AdminAccount, error) {
	admin := &models.AdminAccount{}
	query := `
		SELECT id, username, passwordHash, email, mobile, status, isSuperAdmin, 
		       realName, avatar, departmentId, note, lastLoginIp, lastLoginTime,
		       createdBy, updatedBy, createdTime, updatedTime 
		FROM adminAccount 
		WHERE username = ?
	`
	
	row := db.MySQLDBGameWeb.QueryRow(query, username)
	
	var departmentId, createdBy, updatedBy sql.NullInt64
	var lastLoginTime, createdTime, updatedTime sql.NullTime
	var avatar, mobile, note, lastLoginIp sql.NullString

	err := row.Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Email,
		&mobile, &admin.Status, &admin.IsSuperAdmin, &admin.RealName,
		&avatar, &departmentId, &note, &lastLoginIp, &lastLoginTime,
		&createdBy, &updatedBy, &createdTime, &updatedTime,
	)

	if err != nil {
		return nil, err
	}

	// 处理NULL值
	if mobile.Valid {
		admin.Mobile = mobile.String
	}
	if avatar.Valid {
		admin.Avatar = avatar.String
	}
	if departmentId.Valid {
		deptId := int(departmentId.Int64)
		admin.DepartmentID = &deptId
	}
	if note.Valid {
		admin.Note = note.String
	}
	if lastLoginIp.Valid {
		admin.LastLoginIP = lastLoginIp.String
	}
	if lastLoginTime.Valid {
		admin.LastLoginTime = lastLoginTime.Time
	}
	if createdBy.Valid {
		createdByID := uint64(createdBy.Int64)
		admin.CreatedBy = &createdByID
	}
	if updatedBy.Valid {
		updatedByID := uint64(updatedBy.Int64)
		admin.UpdatedBy = &updatedByID
	}
	if createdTime.Valid {
		admin.CreatedTime = createdTime.Time
	}
	if updatedTime.Valid {
		admin.UpdatedTime = updatedTime.Time
	}

	return admin, nil
}

// getAdminByID 根据ID查询管理员
func getAdminByID(id uint64) (*models.AdminAccount, error) {
	admin := &models.AdminAccount{}
	query := `
		SELECT id, username, passwordHash, email, mobile, status, isSuperAdmin, 
		       realName, avatar, departmentId, note, lastLoginIp, lastLoginTime,
		       createdBy, updatedBy, createdTime, updatedTime 
		FROM adminAccount 
		WHERE id = ?
	`
	
	row := db.MySQLDBGameWeb.QueryRow(query, id)
	
	var departmentId, createdBy, updatedBy sql.NullInt64
	var lastLoginTime, createdTime, updatedTime sql.NullTime
	var avatar, mobile, note, lastLoginIp sql.NullString

	err := row.Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Email,
		&mobile, &admin.Status, &admin.IsSuperAdmin, &admin.RealName,
		&avatar, &departmentId, &note, &lastLoginIp, &lastLoginTime,
		&createdBy, &updatedBy, &createdTime, &updatedTime,
	)

	if err != nil {
		return nil, err
	}

	// 处理NULL值（与getAdminByUsername相同的逻辑）
	if mobile.Valid {
		admin.Mobile = mobile.String
	}
	if avatar.Valid {
		admin.Avatar = avatar.String
	}
	if departmentId.Valid {
		deptId := int(departmentId.Int64)
		admin.DepartmentID = &deptId
	}
	if note.Valid {
		admin.Note = note.String
	}
	if lastLoginIp.Valid {
		admin.LastLoginIP = lastLoginIp.String
	}
	if lastLoginTime.Valid {
		admin.LastLoginTime = lastLoginTime.Time
	}
	if createdBy.Valid {
		createdByID := uint64(createdBy.Int64)
		admin.CreatedBy = &createdByID
	}
	if updatedBy.Valid {
		updatedByID := uint64(updatedBy.Int64)
		admin.UpdatedBy = &updatedByID
	}
	if createdTime.Valid {
		admin.CreatedTime = createdTime.Time
	}
	if updatedTime.Valid {
		admin.UpdatedTime = updatedTime.Time
	}

	return admin, nil
}

// updateAdminLoginInfo 更新管理员登录信息
func updateAdminLoginInfo(admin *models.AdminAccount) error {
	query := `
		UPDATE adminAccount 
		SET lastLoginIp = ?, lastLoginTime = ?, updatedTime = CURRENT_TIMESTAMP 
		WHERE id = ?
	`
	
	_, err := db.MySQLDBGameWeb.Exec(query, admin.LastLoginIP, admin.LastLoginTime, admin.ID)
	return err
}

// checkAdminUsernameExists 检查用户名是否存在
func checkAdminUsernameExists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM adminAccount WHERE username = ?"
	err := db.MySQLDBGameWeb.QueryRow(query, username).Scan(&count)
	return count > 0, err
}

// checkAdminEmailExists 检查邮箱是否存在
func checkAdminEmailExists(email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM adminAccount WHERE email = ?"
	err := db.MySQLDBGameWeb.QueryRow(query, email).Scan(&count)
	return count > 0, err
}

// createAdmin 创建管理员
func createAdmin(admin *models.AdminAccount) error {
	query := `
		INSERT INTO adminAccount (
			username, passwordHash, email, mobile, status, isSuperAdmin,
			realName, avatar, departmentId, note, createdBy, createdTime, updatedTime
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := db.MySQLDBGameWeb.Exec(
		query,
		admin.Username, admin.PasswordHash, admin.Email, admin.Mobile,
		admin.Status, admin.IsSuperAdmin, admin.RealName, admin.Avatar,
		admin.DepartmentID, admin.Note, admin.CreatedBy,
		admin.CreatedTime, admin.UpdatedTime,
	)
	
	if err != nil {
		return err
	}
	
	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	admin.ID = uint64(id)
	return nil
}

// generateSecureToken 生成安全令牌
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// 使用hex编码
	return fmt.Sprintf("%x", bytes), nil
}