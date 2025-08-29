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
	"strconv"
	"strings"
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
		Mobile:        admin.Mobile,
		RealName:      admin.RealName,
		Avatar:        admin.Avatar,
		DepartmentID:  admin.DepartmentID,
		Note:          admin.Note,
		Status:        admin.Status,
		IsSuperAdmin:  admin.IsSuperAdmin == 1,
		LastLoginIP:   admin.LastLoginIP,
		LastLoginTime: admin.LastLoginTime,
		CreatedBy:     admin.CreatedBy,
		UpdatedBy:     admin.UpdatedBy,
		CreatedTime:   admin.CreatedTime,
		UpdatedTime:   admin.UpdatedTime,
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
		Mobile:        admin.Mobile,
		RealName:      admin.RealName,
		Avatar:        admin.Avatar,
		DepartmentID:  admin.DepartmentID,
		Note:          admin.Note,
		Status:        admin.Status,
		IsSuperAdmin:  admin.IsSuperAdmin == 1,
		LastLoginIP:   admin.LastLoginIP,
		LastLoginTime: admin.LastLoginTime,
		CreatedBy:     admin.CreatedBy,
		UpdatedBy:     admin.UpdatedBy,
		CreatedTime:   admin.CreatedTime,
		UpdatedTime:   admin.UpdatedTime,
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

// checkAdminEmailExistsExcludeID 检查邮箱是否存在（排除指定ID）
func checkAdminEmailExistsExcludeID(email string, excludeID uint64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM adminAccount WHERE email = ? AND id != ?"
	err := db.MySQLDBGameWeb.QueryRow(query, email, excludeID).Scan(&count)
	return count > 0, err
}

// updateAdminInfo 更新管理员信息
func updateAdminInfo(adminID uint64, email, mobile, realName, avatar *string, departmentID *int, note *string, updatedBy uint64) error {
	// 构建动态SQL语句
	var setParts []string
	var args []interface{}

	if email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *email)
	}
	if mobile != nil {
		setParts = append(setParts, "mobile = ?")
		args = append(args, *mobile)
	}
	if realName != nil {
		setParts = append(setParts, "realName = ?")
		args = append(args, *realName)
	}
	if avatar != nil {
		setParts = append(setParts, "avatar = ?")
		args = append(args, *avatar)
	}
	if departmentID != nil {
		setParts = append(setParts, "departmentId = ?")
		args = append(args, *departmentID)
	}
	if note != nil {
		setParts = append(setParts, "note = ?")
		args = append(args, *note)
	}

	// 如果没有任何字段需要更新，直接返回
	if len(setParts) == 0 {
		return nil
	}

	// 添加固定的更新字段
	setParts = append(setParts, "updatedBy = ?", "updatedTime = CURRENT_TIMESTAMP")
	args = append(args, updatedBy)

	// 添加WHERE条件
	args = append(args, adminID)

	// 构建完整的SQL语句
	query := fmt.Sprintf("UPDATE adminAccount SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := db.MySQLDBGameWeb.Exec(query, args...)
	return err
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

// UpdateAdmin 更新管理员信息
func UpdateAdmin(c *gin.Context) {
	// 获取路径参数中的管理员ID
	adminID := c.Param("id")
	if adminID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "管理员ID不能为空",
		})
		return
	}

	// 解析管理员ID
	var targetAdminID uint64
	if _, err := fmt.Sscanf(adminID, "%d", &targetAdminID); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的管理员ID格式",
		})
		return
	}

	// 获取当前操作的管理员ID
	currentAdminID, exists := c.Get("adminId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	// 获取当前管理员信息以检查权限
	currentAdmin, err := getAdminByID(currentAdminID.(uint64))
	if err != nil {
		log.Errorf("查询当前管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 权限检查：只能修改自己的信息，或者超级管理员可以修改任何人的信息
	if currentAdminID.(uint64) != targetAdminID && currentAdmin.IsSuperAdmin != 1 {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Code:    403,
			Message: "没有权限修改该管理员信息",
		})
		return
	}

	// 绑定请求参数
	var req struct {
		Email        *string `json:"email" binding:"omitempty,email,max=100"`
		Mobile       *string `json:"mobile" binding:"omitempty,max=20"`
		RealName     *string `json:"realName" binding:"omitempty,min=1,max=50"`
		Avatar       *string `json:"avatar" binding:"omitempty,max=255"`
		DepartmentID *int    `json:"departmentId"`
		Note         *string `json:"note" binding:"omitempty,max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("更新管理员参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 检查要更新的管理员是否存在
	targetAdmin, err := getAdminByID(targetAdminID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Code:    404,
				Message: "管理员不存在",
			})
			return
		}
		log.Errorf("查询目标管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 检查邮箱唯一性（如果要更新邮箱）
	if req.Email != nil && *req.Email != targetAdmin.Email {
		if exists, err := checkAdminEmailExistsExcludeID(*req.Email, targetAdminID); err != nil {
			log.Errorf("检查邮箱唯一性失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "系统错误",
			})
			return
		} else if exists {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "邮箱已被其他管理员使用",
			})
			return
		}
	}

	// 执行更新操作
	if err := updateAdminInfo(targetAdminID, req.Email, req.Mobile, req.RealName, req.Avatar, req.DepartmentID, req.Note, currentAdminID.(uint64)); err != nil {
		log.Errorf("更新管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	log.Infof("管理员信息更新成功: ID: %d, 操作者: %d", targetAdminID, currentAdminID.(uint64))
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "更新成功",
	})
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

// DeleteAdmin 删除管理员账户（仅超级管理员可用）
func DeleteAdmin(c *gin.Context) {
	// 获取路径参数中的管理员ID
	adminID := c.Param("id")
	if adminID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "管理员ID不能为空",
		})
		return
	}

	// 解析管理员ID
	var targetAdminID uint64
	if _, err := fmt.Sscanf(adminID, "%d", &targetAdminID); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的管理员ID格式",
		})
		return
	}

	// 获取当前操作的管理员信息
	currentAdminID, exists := c.Get("adminId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	// 获取当前管理员信息以检查权限
	currentAdmin, err := getAdminByID(currentAdminID.(uint64))
	if err != nil {
		log.Errorf("查询当前管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 权限检查：只有超级管理员才能删除其他管理员
	if currentAdmin.IsSuperAdmin != 1 {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Code:    403,
			Message: "仅超级管理员可执行此操作",
		})
		return
	}

	// 不能删除自己的账户
	if currentAdminID.(uint64) == targetAdminID {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "不能删除自己的账户",
		})
		return
	}

	// 检查要删除的管理员是否存在
	targetAdmin, err := getAdminByID(targetAdminID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Code:    404,
				Message: "管理员不存在",
			})
			return
		}
		log.Errorf("查询目标管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 检查是否为最后一个超级管理员
	if targetAdmin.IsSuperAdmin == 1 {
		superAdminCount, err := countSuperAdmins()
		if err != nil {
			log.Errorf("统计超级管理员数量失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "系统错误",
			})
			return
		}
		
		if superAdminCount <= 1 {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "不能删除最后一个超级管理员账户",
			})
			return
		}
	}

	// 执行删除操作
	if err := deleteAdminByID(targetAdminID); err != nil {
		log.Errorf("删除管理员失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "删除失败",
		})
		return
	}

	// 清除相关的Redis会话
	sessionKey := fmt.Sprintf("admin_session:%d", targetAdminID)
	if err := db.DelRedis(sessionKey); err != nil {
		log.Errorf("清除管理员会话失败: %v", err)
		// 不影响主流程，只记录日志
	}

	log.Infof("管理员账户删除成功: ID: %d (%s), 操作者: %d (%s)", 
		targetAdminID, targetAdmin.Username, currentAdminID.(uint64), currentAdmin.Username)
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "删除成功",
	})
}

// countSuperAdmins 统计超级管理员数量
func countSuperAdmins() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM adminAccount WHERE isSuperAdmin = 1 AND status = 1"
	err := db.MySQLDBGameWeb.QueryRow(query).Scan(&count)
	return count, err
}

// deleteAdminByID 根据ID删除管理员
func deleteAdminByID(adminID uint64) error {
	query := "DELETE FROM adminAccount WHERE id = ?"
	_, err := db.MySQLDBGameWeb.Exec(query, adminID)
	return err
}

// GetAdminList 获取管理员列表（仅超级管理员可用）
func GetAdminList(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	isSuperAdmin := c.Query("isSuperAdmin")

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取当前管理员信息以检查权限
	currentAdminID, exists := c.Get("adminId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	currentAdmin, err := getAdminByID(currentAdminID.(uint64))
	if err != nil {
		log.Errorf("查询当前管理员信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 权限检查：只有超级管理员可以查看所有管理员列表
	if currentAdmin.IsSuperAdmin != 1 {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Code:    403,
			Message: "仅超级管理员可查看管理员列表",
		})
		return
	}

	// 查询管理员列表
	admins, total, err := getAdminList(page, pageSize, keyword, status, isSuperAdmin)
	if err != nil {
		log.Errorf("查询管理员列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "查询失败",
		})
		return
	}

	log.Infof("管理员查询管理员列表: 管理员ID=%v, 页码=%d, 页大小=%d, 关键词=%s, 总数=%d", 
		currentAdminID, page, pageSize, keyword, total)

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "查询成功",
		Data: gin.H{
			"list":     admins,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// getAdminList 查询管理员列表
func getAdminList(page, pageSize int, keyword, status, isSuperAdmin string) ([]*models.AdminInfo, int64, error) {
	// 构建查询条件
	var whereConditions []string
	var args []interface{}

	// 基本条件：不查询已删除的管理员（可以根据实际情况调整）
	// whereConditions = append(whereConditions, "status >= 0")

	// 关键词搜索（用户名、邮箱、真实姓名）
	if keyword != "" {
		whereConditions = append(whereConditions, "(username LIKE ? OR email LIKE ? OR realName LIKE ?)")
		likeKeyword := "%" + keyword + "%"
		args = append(args, likeKeyword, likeKeyword, likeKeyword)
	}

	// 状态筛选
	if status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, status)
	}

	// 超级管理员筛选
	if isSuperAdmin != "" {
		whereConditions = append(whereConditions, "isSuperAdmin = ?")
		args = append(args, isSuperAdmin)
	}

	// 构建 WHERE 子句
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM adminAccount
		%s
	`, whereClause)

	var total int64
	err := db.MySQLDBGameWeb.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	listQuery := fmt.Sprintf(`
		SELECT id, username, email, mobile, realName, avatar, departmentId, note, 
		       status, isSuperAdmin, lastLoginIp, lastLoginTime, 
		       createdBy, updatedBy, createdTime, updatedTime
		FROM adminAccount
		%s
		ORDER BY createdTime DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数
	finalArgs := append(args, pageSize, offset)

	rows, err := db.MySQLDBGameWeb.Query(listQuery, finalArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var admins []*models.AdminInfo
	for rows.Next() {
		var admin models.AdminAccount
		var departmentId, createdBy, updatedBy sql.NullInt64
		var lastLoginTime, createdTime, updatedTime sql.NullTime
		var avatar, mobile, note, lastLoginIp sql.NullString

		err := rows.Scan(
			&admin.ID, &admin.Username, &admin.Email, &mobile,
			&admin.RealName, &avatar, &departmentId, &note,
			&admin.Status, &admin.IsSuperAdmin, &lastLoginIp, &lastLoginTime,
			&createdBy, &updatedBy, &createdTime, &updatedTime,
		)
		if err != nil {
			return nil, 0, err
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

		// 转换为 AdminInfo 类型
		adminInfo := &models.AdminInfo{
			ID:            admin.ID,
			Username:      admin.Username,
			Email:         admin.Email,
			Mobile:        admin.Mobile,
			RealName:      admin.RealName,
			Avatar:        admin.Avatar,
			DepartmentID:  admin.DepartmentID,
			Note:          admin.Note,
			Status:        admin.Status,
			IsSuperAdmin:  admin.IsSuperAdmin == 1,
			LastLoginIP:   admin.LastLoginIP,
			LastLoginTime: admin.LastLoginTime,
			CreatedBy:     admin.CreatedBy,
			UpdatedBy:     admin.UpdatedBy,
			CreatedTime:   admin.CreatedTime,
			UpdatedTime:   admin.UpdatedTime,
		}

		admins = append(admins, adminInfo)
	}

	return admins, total, nil
}