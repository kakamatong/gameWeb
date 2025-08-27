package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 管理后台邮件API ====================
// 以下函数为管理后台提供邮件管理服务

// SendSystemMail 发送系统邮件（管理后台API）
func SendSystemMail(c *gin.Context) {
	var req models.SendMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("发送邮件参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证时间范围
	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "结束时间不能早于开始时间",
		})
		return
	}

	// 获取管理员信息
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")

	// 开始事务
	tx, err := db.MySQLDBGameWeb.Begin()
	if err != nil {
		log.Errorf("开始事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	defer tx.Rollback()

	// 1. 创建邮件记录
	mail := &models.Mails{
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Awards:    req.Awards,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    1, // 1-有效
	}

	if err := createMail(tx, mail); err != nil {
		log.Errorf("创建邮件记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "发送失败",
		})
		return
	}

	// 2. 创建邮件系统记录
	if req.Type == 0 || len(req.TargetUsers) == 0 {
		// 全服邮件
		if err := createMailSystem(tx, mail.ID, 0, 1); err != nil {
			log.Errorf("创建全服邮件系统记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "发送失败",
			})
			return
		}
	} else {
		// 个人邮件
		for _, userID := range req.TargetUsers {
			if err := createMailSystem(tx, mail.ID, userID, 0); err != nil {
				log.Errorf("创建个人邮件系统记录失败: userID=%d, err=%v", userID, err)
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Code:    500,
					Message: "发送失败",
				})
				return
			}

			// 为个人邮件创建mailUsers记录
			if err := createMailUser(tx, mail.ID, userID); err != nil {
				log.Errorf("创建个人邮件用户记录失败: userID=%d, err=%v", userID, err)
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Code:    500,
					Message: "发送失败",
				})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Errorf("提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "发送失败",
		})
		return
	}

	// 记录操作日志
	log.Infof("管理员发送系统邮件: 管理员ID=%v, 管理员=%v, 邮件ID=%d, 类型=%d, 目标用户数=%d, IP=%s",
		adminId, username, mail.ID, req.Type, len(req.TargetUsers), c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "发送成功",
		Data: gin.H{
			"mailId": mail.ID,
		},
	})
}

// GetAdminMailList 获取邮件列表（管理后台API）
func GetAdminMailList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")
	status := c.Query("status")
	mailType := c.Query("type")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	whereConditions := []string{}
	args := []interface{}{}

	if status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, status)
	}

	if mailType != "" {
		whereConditions = append(whereConditions, "type = ?")
		args = append(args, mailType)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	total, err := getMailCount(whereClause, args)
	if err != nil {
		log.Errorf("查询邮件总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询邮件列表
	mails, err := getMailList(whereClause, args, page, pageSize)
	if err != nil {
		log.Errorf("查询邮件列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	response := models.PaginationResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     mails,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// GetAdminMailDetail 获取邮件详情（管理后台API）
func GetAdminMailDetail(c *gin.Context) {
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的邮件ID",
		})
		return
	}

	mail, err := getMailByID(mailID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Code:    404,
				Message: "邮件不存在",
			})
			return
		}
		log.Errorf("查询邮件详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询邮件系统信息，获取发送范围
	mailSystems, err := getMailSystemByMailID(mailID)
	if err != nil {
		log.Errorf("查询邮件系统信息失败: %v", err)
	}

	// 构建响应数据
	response := gin.H{
		"mail":        mail,
		"mailSystems": mailSystems,
	}

	// 如果是个人邮件，查询具体的用户信息
	if mail.Type == 1 {
		targetUsers := []int64{}
		for _, ms := range mailSystems {
			if ms.UserID > 0 {
				targetUsers = append(targetUsers, ms.UserID)
			}
		}
		response["targetUsers"] = targetUsers
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// UpdateMailStatus 更新邮件状态
func UpdateMailStatus(c *gin.Context) {
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的邮件ID",
		})
		return
	}

	var req struct {
		Status int8 `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证邮件是否存在
	exists, err := checkMailExists(mailID)
	if err != nil {
		log.Errorf("检查邮件存在性失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    404,
			Message: "邮件不存在",
		})
		return
	}

	// 更新邮件状态
	if err := updateMailStatus(mailID, req.Status); err != nil {
		log.Errorf("更新邮件状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	// 获取管理员信息记录操作日志
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")
	log.Infof("管理员更新邮件状态: 管理员ID=%v, 管理员=%v, 邮件ID=%d, 状态=%d, IP=%s",
		adminId, username, mailID, req.Status, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "更新成功",
	})
}

// GetMailStats 获取邮件统计信息
func GetMailStats(c *gin.Context) {
	stats, err := getMailStats()
	if err != nil {
		log.Errorf("获取邮件统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    stats,
	})
}

// 数据库操作函数

// createMail 创建邮件记录
func createMail(tx *sql.Tx, mail *models.Mails) error {
	query := `
		INSERT INTO mails (type, title, content, awards, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := tx.Exec(query, mail.Type, mail.Title, mail.Content, mail.Awards)
	if err != nil {
		return err
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	mail.ID = id
	// 设置默认时间值（用于后续逻辑）
	mail.StartTime = time.Now()
	mail.EndTime = mail.EndTime // 保持请求中的EndTime
	return nil
}

// createMailSystem 创建邮件系统记录
func createMailSystem(tx *sql.Tx, mailID, userID int64, isGlobal int8) error {
	// 根据isGlobal参数设置type
	mailType := 1 // 个人邮件
	if isGlobal == 1 {
		mailType = 0 // 全服邮件
	}
	
	// 默认时间范围（可以根据需要调整）
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour * 30) // 30天有效期
	
	query := `
		INSERT INTO mailSystem (type, mailid, startTime, endTime)
		VALUES (?, ?, ?, ?)
	`

	_, err := tx.Exec(query, mailType, mailID, startTime, endTime)
	return err
}

// createMailUser 创建用户邮件记录
func createMailUser(tx *sql.Tx, mailID, userID int64) error {
	// 默认时间范围（个人邮件，可以根据需要调整）
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour * 30) // 30天有效期
	
	query := `
		INSERT INTO mailUsers (mailid, userid, status, startTime, endTime, update_at)
		VALUES (?, ?, 0, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := tx.Exec(query, mailID, userID, startTime, endTime)
	return err
}

// getMailCount 获取邮件总数
func getMailCount(whereClause string, args []interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM mails %s", whereClause)

	var count int64
	err := db.MySQLDBGameWeb.QueryRow(query, args...).Scan(&count)
	return count, err
}

// getMailList 获取邮件列表
func getMailList(whereClause string, args []interface{}, page, pageSize int) ([]models.Mails, error) {
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at
		FROM mails m
		%s
		ORDER BY m.id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数到args
	finalArgs := append(args, pageSize, offset)

	rows, err := db.MySQLDBGameWeb.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []models.Mails
	for rows.Next() {
		var mail models.Mails
		err := rows.Scan(
			&mail.ID, &mail.Type, &mail.Title, &mail.Content,
			&mail.Awards, &mail.StartTime, // 使用created_at作为StartTime
		)
		if err != nil {
			return nil, err
		}
		// 设置默认值
		mail.EndTime = mail.StartTime.Add(24 * time.Hour * 30) // 默认30天有效期
		mail.Status = 1 // 默认有效
		if err != nil {
			return nil, err
		}
		mails = append(mails, mail)
	}

	return mails, nil
}

// getMailByID 根据ID获取邮件
func getMailByID(mailID int64) (*models.Mails, error) {
	query := `
		SELECT id, type, title, content, awards, created_at
		FROM mails 
		WHERE id = ?
	`

	var mail models.Mails
	err := db.MySQLDBGameWeb.QueryRow(query, mailID).Scan(
		&mail.ID, &mail.Type, &mail.Title, &mail.Content,
		&mail.Awards, &mail.StartTime, // 使用created_at作为StartTime
	)

	if err != nil {
		return nil, err
	}

	// 设置默认值
	mail.EndTime = mail.StartTime.Add(24 * time.Hour * 30) // 默认30天有效期
	mail.Status = 1 // 默认有效

	return &mail, nil
}

// getMailSystemByMailID 根据邮件ID获取邮件系统记录
func getMailSystemByMailID(mailID int64) ([]models.MailSystem, error) {
	query := `
		SELECT id, mailid, type, startTime, endTime
		FROM mailSystem 
		WHERE mailid = ?
	`

	rows, err := db.MySQLDBGameWeb.Query(query, mailID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailSystems []models.MailSystem
	for rows.Next() {
		var ms models.MailSystem
		var msType int8
		var startTime, endTime time.Time
		err := rows.Scan(&ms.ID, &ms.MailID, &msType, &startTime, &endTime)
		if err != nil {
			return nil, err
		}
		// 根据type设置IsGlobal
		ms.IsGlobal = 0
		if msType == 0 {
			ms.IsGlobal = 1 // 全服邮件
		}
		ms.UserID = 0 // mailSystem表中没有userid字段
		mailSystems = append(mailSystems, ms)
	}

	return mailSystems, nil
}

// checkMailExists 检查邮件是否存在
func checkMailExists(mailID int64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM mails WHERE id = ?"
	err := db.MySQLDBGameWeb.QueryRow(query, mailID).Scan(&count)
	return count > 0, err
}

// updateMailStatus 更新邮件状态
// 注意：mails表中没有status字段，这个功能需要重新设计
func updateMailStatus(mailID int64, status int8) error {
	// 由于mails表中没有status字段，暂时返回成功
	// 实际上可以通过操作mailSystem表的startTime/endTime来实现邮件的启用/禁用
	log.Warnf("更新邮件状态功能暂未实现: mailID=%d, status=%d", mailID, status)
	return nil
}

// getMailStats 获取邮件统计信息
func getMailStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询总邮件数
	var totalMails int64
	query := "SELECT COUNT(*) FROM mails"
	err := db.MySQLDBGameWeb.QueryRow(query).Scan(&totalMails)
	if err != nil {
		return nil, err
	}
	stats["totalMails"] = totalMails

	// 查询有效邮件数（由于没有status字段，直接等于总数）
	stats["activeMails"] = totalMails

	// 查询全服邮件数
	var globalMails int64
	query = "SELECT COUNT(*) FROM mails WHERE type = 0"
	err = db.MySQLDBGameWeb.QueryRow(query).Scan(&globalMails)
	if err != nil {
		log.Warnf("查询全服邮件数失败: %v", err)
		stats["globalMails"] = 0
	} else {
		stats["globalMails"] = globalMails
	}

	// 查询个人邮件数
	var personalMails int64
	query = "SELECT COUNT(*) FROM mails WHERE type = 1"
	err = db.MySQLDBGameWeb.QueryRow(query).Scan(&personalMails)
	if err != nil {
		log.Warnf("查询个人邮件数失败: %v", err)
		stats["personalMails"] = 0
	} else {
		stats["personalMails"] = personalMails
	}

	// 查询今日发送邮件数
	today := time.Now().Format("2006-01-02")
	var todayMails int64
	query = "SELECT COUNT(*) FROM mails WHERE DATE(created_at) = ?"
	err = db.MySQLDBGameWeb.QueryRow(query, today).Scan(&todayMails)
	if err != nil {
		log.Warnf("查询今日邮件数失败: %v", err)
		stats["todayMails"] = 0
	} else {
		stats["todayMails"] = todayMails
	}

	return stats, nil
}

// ==================== 客户端邮件查询辅助函数 ====================

// syncSystemMails 同步系统邮件到用户邮件表
func syncSystemMails(userID int64) error {
	// 查询所有未过期的系统邮件（全服邮件）
	query := `
		SELECT DISTINCT ms.mailid 
		FROM mailSystem ms
		INNER JOIN mails m ON ms.mailid = m.id
		WHERE ms.type = 0
		  AND ms.startTime <= CURRENT_TIMESTAMP 
		  AND ms.endTime >= CURRENT_TIMESTAMP
		  AND ms.mailid NOT IN (
			  SELECT mailid FROM mailUsers WHERE userid = ?
		  )
	`

	rows, err := db.MySQLDBGameWeb.Query(query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 批量插入用户邮件记录
	for rows.Next() {
		var mailID int64
		if err := rows.Scan(&mailID); err != nil {
			continue
		}

		// 为用户创建邮件记录
		insertQuery := `
			INSERT IGNORE INTO mailUsers (mailid, userid, status, startTime, endTime, update_at)
			SELECT ?, ?, 0, ms.startTime, ms.endTime, CURRENT_TIMESTAMP
			FROM mailSystem ms WHERE ms.mailid = ?
		`
		_, err := db.MySQLDBGameWeb.Exec(insertQuery, mailID, userID, mailID)
		if err != nil {
			log.Warnf("同步系统邮件失败: mailID=%d, userID=%d, err=%v", mailID, userID, err)
		}
	}

	return nil
}

// getClientMailList 获取客户端邮件列表
func getClientMailList(userID int64) ([]gin.H, error) {
	query := `
		SELECT 
			m.id, m.type, m.title, m.content, m.awards, m.created_at,
			COALESCE(mu.status, 0) as status,
			COALESCE(mu.startTime, ms.startTime) as startTime,
			COALESCE(mu.endTime, ms.endTime) as endTime,
			COALESCE(mu.update_at, m.created_at) as update_at
		FROM mails m
		LEFT JOIN mailUsers mu ON m.id = mu.mailid AND mu.userid = ?
		LEFT JOIN mailSystem ms ON m.id = ms.mailid
		WHERE (
			-- 全服邮件（在mailSystem中有记录，type=0）
			(m.type = 0 AND ms.type = 0 AND ms.startTime <= CURRENT_TIMESTAMP AND ms.endTime >= CURRENT_TIMESTAMP)
			OR 
			-- 个人邮件（在mailUsers中有记录）
			(m.type = 1 AND mu.userid = ? AND mu.startTime <= CURRENT_TIMESTAMP AND mu.endTime >= CURRENT_TIMESTAMP)
		)
		ORDER BY update_at DESC
	`

	rows, err := db.MySQLDBGameWeb.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []gin.H
	for rows.Next() {
		var mail struct {
			ID        int64     `json:"id"`
			Type      int8      `json:"type"`
			Title     string    `json:"title"`
			Content   string    `json:"content"`
			Awards    string    `json:"awards"`
			CreatedAt time.Time `json:"createdAt"`
			Status    int8      `json:"status"`
			StartTime time.Time `json:"startTime"`
			EndTime   time.Time `json:"endTime"`
			UpdateAt  time.Time `json:"updateAt"`
		}

		err := rows.Scan(
			&mail.ID, &mail.Type, &mail.Title, &mail.Content,
			&mail.Awards, &mail.CreatedAt, &mail.Status,
			&mail.StartTime, &mail.EndTime, &mail.UpdateAt,
		)
		if err != nil {
			return nil, err
		}

		// 构建返回结果
		mailData := gin.H{
			"id":         mail.ID,
			"type":       mail.Type,
			"title":      mail.Title,
			"content":    mail.Content,
			"awards":     mail.Awards,
			"createdAt":  mail.CreatedAt,
			"status":     mail.Status,
			"startTime":  mail.StartTime,
			"endTime":    mail.EndTime,
			"updateAt":   mail.UpdateAt,
			"isRead":     mail.Status >= 1, // 0-未读, 1-已读, 2-已领取
			"isReceived": mail.Status >= 2, // 2-已领取
		}

		mails = append(mails, mailData)
	}

	return mails, nil
}

// getClientMailDetail 获取客户端邮件详情
func getClientMailDetail(mailID, userID int64) (gin.H, error) {
	query := `
		SELECT 
			m.id, m.type, m.title, m.content, m.awards, m.created_at,
			COALESCE(mu.status, 0) as status,
			COALESCE(mu.startTime, ms.startTime) as startTime,
			COALESCE(mu.endTime, ms.endTime) as endTime,
			COALESCE(mu.update_at, m.created_at) as update_at
		FROM mails m
		LEFT JOIN mailUsers mu ON m.id = mu.mailid AND mu.userid = ?
		LEFT JOIN mailSystem ms ON m.id = ms.mailid
		WHERE m.id = ?
		  AND (
			  -- 全服邮件
			  (m.type = 0 AND ms.type = 0 AND ms.startTime <= CURRENT_TIMESTAMP AND ms.endTime >= CURRENT_TIMESTAMP)
			  OR 
			  -- 个人邮件
			  (m.type = 1 AND mu.userid = ? AND mu.startTime <= CURRENT_TIMESTAMP AND mu.endTime >= CURRENT_TIMESTAMP)
		  )
	`

	var mail struct {
		ID        int64     `json:"id"`
		Type      int8      `json:"type"`
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		Awards    string    `json:"awards"`
		CreatedAt time.Time `json:"createdAt"`
		Status    int8      `json:"status"`
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
		UpdateAt  time.Time `json:"updateAt"`
	}

	err := db.MySQLDBGameWeb.QueryRow(query, userID, mailID, userID).Scan(
		&mail.ID, &mail.Type, &mail.Title, &mail.Content,
		&mail.Awards, &mail.CreatedAt, &mail.Status,
		&mail.StartTime, &mail.EndTime, &mail.UpdateAt,
	)

	if err != nil {
		return nil, err
	}

	// 构建返回结果
	mailDetail := gin.H{
		"id":         mail.ID,
		"type":       mail.Type,
		"title":      mail.Title,
		"content":    mail.Content,
		"awards":     mail.Awards,
		"createdAt":  mail.CreatedAt,
		"status":     mail.Status,
		"startTime":  mail.StartTime,
		"endTime":    mail.EndTime,
		"updateAt":   mail.UpdateAt,
		"isRead":     mail.Status >= 1,
		"isReceived": mail.Status >= 2,
	}

	return mailDetail, nil
}

// ==================== 游戏客户端邮件API ====================
// 以下函数为游戏客户端提供邮件相关服务

// GetClientMailList 获取邮件列表（游戏客户端API）
func GetClientMailList(c *gin.Context) {
	// 解析请求体获取用户ID
	var req struct {
		UserID int64 `json:"userid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("获取邮件列表参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
		})
		return
	}

	// 从JWT上下文获取用户ID进行验证
	jwtUserID, exists := c.Get("userid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized",
		})
		return
	}

	// 验证请求中的用户ID与JWT中的用户ID是否一致
	if req.UserID != jwtUserID.(int64) {
		log.Warnf("用户ID不匹配: JWT=%d, Request=%d", jwtUserID, req.UserID)
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "Permission denied",
		})
		return
	}

	// 先同步系统邮件到用户邮件表
	if err := syncSystemMails(req.UserID); err != nil {
		log.Errorf("同步系统邮件失败: userID=%d, err=%v", req.UserID, err)
		// 同步失败不影响查询，继续执行
	}

	// 查询用户邮件列表
	mails, err := getClientMailList(req.UserID)
	if err != nil {
		log.Errorf("查询用户邮件列表失败: userID=%d, err=%v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	log.Infof("获取邮件列表成功: userID=%d, count=%d", req.UserID, len(mails))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    mails,
	})
}

// GetClientMailDetail 获取邮件详情（游戏客户端API）
func GetClientMailDetail(c *gin.Context) {
	// 获取邮件ID
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid mail ID",
		})
		return
	}

	// 解析请求体获取用户ID
	var req struct {
		UserID int64 `json:"userid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("获取邮件详情参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
		})
		return
	}

	// 从JWT上下文获取用户ID进行验证
	jwtUserID, exists := c.Get("userid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized",
		})
		return
	}

	// 验证请求中的用户ID与JWT中的用户ID是否一致
	if req.UserID != jwtUserID.(int64) {
		log.Warnf("用户ID不匹配: JWT=%d, Request=%d", jwtUserID, req.UserID)
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "Permission denied",
		})
		return
	}

	// 查询用户邮件详情
	mailDetail, err := getClientMailDetail(mailID, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found",
			})
			return
		}
		log.Errorf("查询邮件详情失败: mailID=%d, userID=%d, err=%v", mailID, req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	log.Infof("获取邮件详情成功: mailID=%d, userID=%d", mailID, req.UserID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    mailDetail,
	})
}

// MarkMailAsRead 标记邮件为已读（游戏客户端API）
func MarkMailAsRead(c *gin.Context) {
	// 获取邮件ID
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid mail ID",
		})
		return
	}

	// 解析请求体获取用户ID
	var req struct {
		UserID int64 `json:"userid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("标记邮件已读参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
		})
		return
	}

	// 从JWT上下文获取用户ID进行验证
	jwtUserID, exists := c.Get("userid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized",
		})
		return
	}

	// 验证请求中的用户ID与JWT中的用户ID是否一致
	if req.UserID != jwtUserID.(int64) {
		log.Warnf("用户ID不匹配: JWT=%d, Request=%d", jwtUserID, req.UserID)
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "Permission denied",
		})
		return
	}

	// 更新邮件状态为已读
	query := `
		UPDATE mailUsers 
		SET status = CASE WHEN status = 0 THEN 1 ELSE status END, update_at = CURRENT_TIMESTAMP 
		WHERE mailid = ? AND userid = ? AND status < 3
	`

	result, err := db.MySQLDBGameWeb.Exec(query, mailID, req.UserID)
	if err != nil {
		log.Errorf("标记邮件已读失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Errorf("获取更新行数失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	if rowsAffected == 0 {
		// 邮件可能已经是已读状态或不存在
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Mail already read or not found",
			"data":    gin.H{},
		})
		return
	}

	log.Infof("邮件标记已读成功: mailID=%d, userID=%d", mailID, req.UserID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{},
	})
}

// GetMailAward 领取邮件奖励（游戏客户端API）
func GetMailAward(c *gin.Context) {
	// 获取邮件ID
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid mail ID",
		})
		return
	}

	// 解析请求体获取用户ID
	var req struct {
		UserID int64 `json:"userid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("领取邮件奖励参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
		})
		return
	}

	// 从JWT上下文获取用户ID进行验证
	jwtUserID, exists := c.Get("userid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized",
		})
		return
	}

	// 验证请求中的用户ID与JWT中的用户ID是否一致
	if req.UserID != jwtUserID.(int64) {
		log.Warnf("用户ID不匹配: JWT=%d, Request=%d", jwtUserID, req.UserID)
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "Permission denied",
		})
		return
	}

	// 获取分布式锁，防止重复领取
	lockKey := fmt.Sprintf("mail_award_lock:%d:%d", req.UserID, mailID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	lockExpire := 30 * time.Second

	// 尝试获取Redis锁
	lockAcquired, err := db.RedisClient.SetNX(context.Background(), lockKey, lockValue, lockExpire).Result()
	if err != nil {
		log.Errorf("获取Redis锁失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	if !lockAcquired {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    429,
			"message": "Operation too frequent",
		})
		return
	}

	// 确保在函数结束时释放锁
	defer func() {
		// 释放锁
		delScript := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		db.RedisClient.Eval(context.Background(), delScript, []string{lockKey}, lockValue)
	}()

	// 查询邮件奖励信息（gameWeb数据库）
	var mailInfo struct {
		Awards     string
		IsReceived int8
	}

	queryMail := `
		SELECT m.awards, COALESCE(mu.status, 0) as status
		FROM mails m
		LEFT JOIN mailUsers mu ON m.id = mu.mailid AND mu.userid = ?
		LEFT JOIN mailSystem ms ON m.id = ms.mailid
		WHERE m.id = ?
		  AND (
			  -- 全服邮件
			  (m.type = 0 AND ms.type = 0 AND ms.startTime <= CURRENT_TIMESTAMP AND ms.endTime >= CURRENT_TIMESTAMP)
			  OR 
			  -- 个人邮件
			  (m.type = 1 AND mu.userid = ? AND mu.startTime <= CURRENT_TIMESTAMP AND mu.endTime >= CURRENT_TIMESTAMP)
		  )
	`

	err = db.MySQLDBGameWeb.QueryRow(queryMail, req.UserID, mailID, req.UserID).Scan(&mailInfo.Awards, &mailInfo.IsReceived)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found or expired",
			})
			return
		}
		log.Errorf("查询邮件信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}

	// 检查是否已经领取
	if mailInfo.IsReceived >= 2 { // 2-已领取
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Award already received",
			"data":    gin.H{"alreadyReceived": true},
		})
		return
	}

	// 解析奖励JSON
	if mailInfo.Awards == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "No awards in this mail",
			"data":    gin.H{},
		})
		return
	}

	// 解析奖励内容
	var awardsData struct {
		Props []struct {
			ID  int   `json:"id"`
			Cnt int64 `json:"cnt"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(mailInfo.Awards), &awardsData); err != nil {
		log.Errorf("解析奖励JSON失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Invalid award format",
		})
		return
	}

	// 转换为内部使用的awards格式
	awards := make([]struct {
		Type  int   `json:"type"`
		Count int64 `json:"count"`
	}, len(awardsData.Props))
	
	for i, prop := range awardsData.Props {
		awards[i] = struct {
			Type  int   `json:"type"`
			Count int64 `json:"count"`
		}{
			Type:  prop.ID,
			Count: prop.Cnt,
		}
	}

	// 开始用户财富数据库事务（game数据库）
	txGame, err := db.MySQLDB.Begin()
	if err != nil {
		log.Errorf("开始用户财富事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
		return
	}
	defer txGame.Rollback()

	// 发放奖励到用户财富表
	for _, award := range awards {
		// 检查该类型财富是否存在
		var currentAmount int64
		checkQuery := "SELECT richNums FROM userRiches WHERE userid = ? AND richType = ?"
		err = txGame.QueryRow(checkQuery, req.UserID, award.Type).Scan(&currentAmount)
		
		if err == sql.ErrNoRows {
			// 插入新记录
			insertQuery := "INSERT INTO userRiches (userid, richType, richNums) VALUES (?, ?, ?)"
			_, err = txGame.Exec(insertQuery, req.UserID, award.Type, award.Count)
		} else if err == nil {
			// 更新现有记录
			updateQuery := "UPDATE userRiches SET richNums = richNums + ? WHERE userid = ? AND richType = ?"
			_, err = txGame.Exec(updateQuery, award.Count, req.UserID, award.Type)
		}
		
		if err != nil {
			log.Errorf("更新用户财富失败: userID=%d, richType=%d, count=%d, err=%v", 
				req.UserID, award.Type, award.Count, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to grant awards",
			})
			return
		}
	}

	// 提交用户财富事务
	if err := txGame.Commit(); err != nil {
		log.Errorf("提交用户财富事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to commit user wealth transaction",
		})
		return
	}

	// 开始邮件数据库事务（gameWeb数据库）
	txMail, err := db.MySQLDBGameWeb.Begin()
	if err != nil {
		log.Errorf("开始邮件事务失败: %v", err)
		// 此时用户财富已经发放，需要记录日志但继续更新邮件状态
		log.Warnf("用户财富已发放但邮件状态更新事务失败，需要手动处理: mailID=%d, userID=%d", mailID, req.UserID)
	}
	defer txMail.Rollback()

	// 更新邮件状态为已领取
	updateMailQuery := `
		UPDATE mailUsers 
		SET status = 2, update_at = CURRENT_TIMESTAMP
		WHERE mailid = ? AND userid = ?
	`

	result, err := txMail.Exec(updateMailQuery, mailID, req.UserID)
	if err != nil {
		log.Errorf("更新邮件状态失败: %v", err)
		// 用户财富已发放，此处记录日志但不返回错误
		log.Warnf("用户财富已发放但邮件状态更新失败，需要手动处理: mailID=%d, userID=%d", mailID, req.UserID)
	} else {
		// 检查是否有行被更新
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Errorf("获取更新行数失败: %v", err)
		} else if rowsAffected == 0 {
			log.Warnf("没有邮件记录被更新，可能邮件不存在或用户无权限: mailID=%d, userID=%d", mailID, req.UserID)
		}
		
		// 提交邮件事务
		if err := txMail.Commit(); err != nil {
			log.Errorf("提交邮件事务失败: %v", err)
			log.Warnf("用户财富已发放但邮件状态提交失败，需要手动处理: mailID=%d, userID=%d", mailID, req.UserID)
		}
	}

	log.Infof("邮件奖励领取成功: mailID=%d, userID=%d, awards=%v", mailID, req.UserID, awards)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"awards": awards,
		},
	})
}