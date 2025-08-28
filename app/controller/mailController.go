package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 游戏客户端邮件API (保持原有接口不变) ====================

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

	// 查询邮件奖励信息（game数据库）
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

	// 发放奖励到用户财富表 - 使用INSERT...ON DUPLICATE KEY UPDATE优化
	// 构建批量插入/更新语句
	if len(awards) > 0 {
		var values []string
		var args []interface{}
		
		for _, award := range awards {
			values = append(values, "(?, ?, ?)")
			args = append(args, req.UserID, award.Type, award.Count)
		}
		
		// 使用INSERT...ON DUPLICATE KEY UPDATE一次性处理所有奖励
		// 如果记录不存在则插入，存在则累加
		batchUpsertQuery := fmt.Sprintf(`
			INSERT INTO userRiches (userid, richType, richNums) 
			VALUES %s
			ON DUPLICATE KEY UPDATE richNums = richNums + VALUES(richNums)
		`, strings.Join(values, ","))
		
		if _, err := txGame.Exec(batchUpsertQuery, args...); err != nil {
			log.Errorf("批量插入/更新用户财富失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to grant awards",
			})
			return
		}
		
		log.Infof("成功批量处理用户财富: userID=%d, 奖励数量=%d", req.UserID, len(awards))
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

	// 向游戏服务器发送奖励通知
	noticeID, err := sendAwardNoticeToGameServer(req.UserID, awards)
	if err != nil {
		log.Errorf("发送奖励通知到游戏服务器失败: %v", err)
		// 奖励已经发放成功，通知失败只记录日志，不影响返回结果
	}

	responseData := gin.H{
		"awards": awards,
	}
	
	// 如果通知成功，添加noticeID到响应
	if noticeID > 0 {
		responseData["noticeid"] = noticeID
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    responseData,
	})
}

// ==================== 管理后台邮件API ====================

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

	// 验证awards格式
	if req.Awards != "" {
		var awards models.AwardsStruct
		if err := json.Unmarshal([]byte(req.Awards), &awards); err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "奖励格式错误",
			})
			return
		}
	}

	// 获取管理员信息
	adminId, _ := c.Get("adminId")

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

	// 创建邮件和系统邮件记录
	mailID, err := createSystemMail(tx, &req, adminId)
	if err != nil {
		log.Errorf("创建系统邮件失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "发送失败",
		})
		return
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

	log.Infof("管理员发送系统邮件: 管理员ID=%v, 邮件ID=%d, 类型=%d, IP=%s",
		adminId, mailID, req.Type, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "发送成功",
		Data:    gin.H{"mailId": mailID},
	})
}

// syncSystemMails 同步系统邮件到用户邮件表
func syncSystemMails(userID int64) error {
	// 获取当前生效的系统邮件
	activeSystemMails, err := getActiveSystemMails()
	if err != nil {
		return err
	}

	// 获取用户已有的邮件
	userMailIDs, err := getUserMailIDs(userID)
	if err != nil {
		return err
	}

	// 找出需要新增的邮件
	newMailIDs := []int64{}
	for _, systemMail := range activeSystemMails {
		found := false
		for _, userMailID := range userMailIDs {
			if userMailID == systemMail.MailID {
				found = true
				break
			}
		}
		if !found {
			newMailIDs = append(newMailIDs, systemMail.MailID)
		}
	}

	// 批量插入新邮件到用户邮件表
	if len(newMailIDs) > 0 {
		return insertUserMails(userID, newMailIDs, activeSystemMails)
	}

	return nil
}

// getClientMailList 获取客户端邮件列表
func getClientMailList(userID int64) ([]models.MailDetailResponse, error) {
	now := time.Now()
	query := `
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       mu.status, mu.startTime, mu.endTime
		FROM mails m
		INNER JOIN mailUsers mu ON m.id = mu.mailid
		WHERE mu.userid = ? AND mu.startTime <= ? AND mu.endTime > ? AND mu.status != 3
		ORDER BY m.created_at DESC
	`

	rows, err := db.MySQLDBGameWeb.Query(query, userID, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []models.MailDetailResponse
	for rows.Next() {
		var mail models.MailDetailResponse
		err := rows.Scan(&mail.ID, &mail.Type, &mail.Title, &mail.Content, 
			&mail.Awards, &mail.CreatedAt, &mail.Status, &mail.StartTime, &mail.EndTime)
		if err != nil {
			return nil, err
		}
		mails = append(mails, mail)
	}

	return mails, nil
}

// getClientMailDetail 获取客户端邮件详情
func getClientMailDetail(mailID, userID int64) (*models.MailDetailResponse, error) {
	now := time.Now()
	query := `
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       mu.status, mu.startTime, mu.endTime
		FROM mails m
		INNER JOIN mailUsers mu ON m.id = mu.mailid
		WHERE m.id = ? AND mu.userid = ? AND mu.startTime <= ? AND mu.endTime > ? AND mu.status != 3
	`

	var mail models.MailDetailResponse
	err := db.MySQLDBGameWeb.QueryRow(query, mailID, userID, now, now).Scan(
		&mail.ID, &mail.Type, &mail.Title, &mail.Content, &mail.Awards, 
		&mail.CreatedAt, &mail.Status, &mail.StartTime, &mail.EndTime)
	
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &mail, nil
}

// sendAwardNoticeToGameServer 向游戏服务器发送奖励通知
func sendAwardNoticeToGameServer(userID int64, awards []struct {
	Type  int   `json:"type"`
	Count int64 `json:"count"`
}) (int64, error) {
	// 构建awardMessage JSON字符串
	richTypes := make([]int, len(awards))
	richNums := make([]int64, len(awards))
	
	for i, award := range awards {
		richTypes[i] = award.Type
		richNums[i] = award.Count
	}
	
	awardMessage := map[string]interface{}{
		"richTypes": richTypes,
		"richNums":  richNums,
	}
	
	awardMessageBytes, err := json.Marshal(awardMessage)
	if err != nil {
		return 0, fmt.Errorf("序列化awardMessage失败: %v", err)
	}
	
	// 构建请求数据
	requestData := map[string]interface{}{
		"userid":       userID,
		"awardMessage": string(awardMessageBytes),
	}
	
	requestBytes, err := json.Marshal(requestData)
	if err != nil {
		return 0, fmt.Errorf("序列化请求数据失败: %v", err)
	}
	
	// 构建请求URL
	gameServerURL := fmt.Sprintf("http://%s:%s/awardnotice", 
		config.AppConfig.GameServer.Host,
		config.AppConfig.GameServer.Port)
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 发送POST请求
	resp, err := client.Post(gameServerURL, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return 0, fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("游戏服务器返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var response struct {
		NoticeID int64 `json:"noticeid"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("解析响应失败: %v", err)
	}
	
	log.Infof("成功发送奖励通知到游戏服务器: userID=%d, noticeID=%d, url=%s", 
		userID, response.NoticeID, gameServerURL)
	
	return response.NoticeID, nil
}

// ==================== 数据库操作函数 ====================

// syncSystemMailsToUser 同步系统邮件到用户邮件表
func syncSystemMailsToUser(userID int64) error {
	// 获取当前生效的系统邮件
	activeSystemMails, err := getActiveSystemMails()
	if err != nil {
		return err
	}

	// 获取用户已有的邮件
	userMailIDs, err := getUserMailIDs(userID)
	if err != nil {
		return err
	}

	// 找出需要新增的邮件
	newMailIDs := []int64{}
	for _, systemMail := range activeSystemMails {
		found := false
		for _, userMailID := range userMailIDs {
			if userMailID == systemMail.MailID {
				found = true
				break
			}
		}
		if !found {
			newMailIDs = append(newMailIDs, systemMail.MailID)
		}
	}

	// 批量插入新邮件到用户邮件表
	if len(newMailIDs) > 0 {
		return insertUserMails(userID, newMailIDs, activeSystemMails)
	}

	return nil
}

// getActiveSystemMails 获取当前生效的系统邮件
func getActiveSystemMails() ([]models.MailSystem, error) {
	now := time.Now()
	query := `
		SELECT id, type, mailid, startTime, endTime 
		FROM mailSystem 
		WHERE startTime <= ? AND endTime > ?
	`
	
	rows, err := db.MySQLDBGameWeb.Query(query, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var systemMails []models.MailSystem
	for rows.Next() {
		var mail models.MailSystem
		err := rows.Scan(&mail.ID, &mail.Type, &mail.MailID, &mail.StartTime, &mail.EndTime)
		if err != nil {
			return nil, err
		}
		systemMails = append(systemMails, mail)
	}

	return systemMails, nil
}

// getUserMailIDs 获取用户已有的邮件ID列表
func getUserMailIDs(userID int64) ([]int64, error) {
	query := "SELECT mailid FROM mailUsers WHERE userid = ?"
	rows, err := db.MySQLDBGameWeb.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailIDs []int64
	for rows.Next() {
		var mailID int64
		if err := rows.Scan(&mailID); err != nil {
			return nil, err
		}
		mailIDs = append(mailIDs, mailID)
	}

	return mailIDs, nil
}

// insertUserMails 批量插入用户邮件
func insertUserMails(userID int64, mailIDs []int64, systemMails []models.MailSystem) error {
	// 创建mailID到systemMail的映射
	mailMap := make(map[int64]models.MailSystem)
	for _, mail := range systemMails {
		mailMap[mail.MailID] = mail
	}

	query := `
		INSERT INTO mailUsers (userid, mailid, status, startTime, endTime) 
		VALUES (?, ?, 0, ?, ?)
	`

	for _, mailID := range mailIDs {
		if systemMail, exists := mailMap[mailID]; exists {
			_, err := db.MySQLDBGameWeb.Exec(query, userID, mailID, systemMail.StartTime, systemMail.EndTime)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getUserActiveMails 获取用户当前生效的邮件列表
func getUserActiveMails(userID int64) ([]models.MailDetailResponse, error) {
	now := time.Now()
	query := `
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       mu.status, mu.startTime, mu.endTime
		FROM mails m
		INNER JOIN mailUsers mu ON m.id = mu.mailid
		WHERE mu.userid = ? AND mu.startTime <= ? AND mu.endTime > ? AND mu.status != 3
		ORDER BY m.created_at DESC
	`

	rows, err := db.MySQLDBGameWeb.Query(query, userID, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []models.MailDetailResponse
	for rows.Next() {
		var mail models.MailDetailResponse
		err := rows.Scan(&mail.ID, &mail.Type, &mail.Title, &mail.Content, 
			&mail.Awards, &mail.CreatedAt, &mail.Status, &mail.StartTime, &mail.EndTime)
		if err != nil {
			return nil, err
		}
		mails = append(mails, mail)
	}

	return mails, nil
}

// getMailDetailForUser 获取用户的邮件详情
func getMailDetailForUser(mailID, userID int64) (*models.MailDetailResponse, error) {
	now := time.Now()
	query := `
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       mu.status, mu.startTime, mu.endTime
		FROM mails m
		INNER JOIN mailUsers mu ON m.id = mu.mailid
		WHERE m.id = ? AND mu.userid = ? AND mu.startTime <= ? AND mu.endTime > ? AND mu.status != 3
	`

	var mail models.MailDetailResponse
	err := db.MySQLDBGameWeb.QueryRow(query, mailID, userID, now, now).Scan(
		&mail.ID, &mail.Type, &mail.Title, &mail.Content, &mail.Awards, 
		&mail.CreatedAt, &mail.Status, &mail.StartTime, &mail.EndTime)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &mail, nil
}

// markMailAsRead 标记邮件为已读
func markMailAsRead(mailID, userID int64) error {
	query := "UPDATE mailUsers SET status = 1, update_at = NOW() WHERE mailid = ? AND userid = ? AND status = 0"
	_, err := db.MySQLDBGameWeb.Exec(query, mailID, userID)
	return err
}

// claimMailAward 领取邮件奖励
func claimMailAward(mailID, userID int64) (string, error) {
	// 检查邮件状态
	var status int8
	var awards string
	query := `
		SELECT mu.status, m.awards
		FROM mailUsers mu
		INNER JOIN mails m ON mu.mailid = m.id
		WHERE mu.mailid = ? AND mu.userid = ? AND mu.endTime > NOW()
	`
	
	err := db.MySQLDBGameWeb.QueryRow(query, mailID, userID).Scan(&status, &awards)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("邮件不存在或已过期")
	}
	if err != nil {
		return "", err
	}

	if status == 2 {
		return "", fmt.Errorf("奖励已经领取过了")
	}
	if status == 3 {
		return "", fmt.Errorf("邮件已删除")
	}

	// 更新邮件状态为已领取
	updateQuery := "UPDATE mailUsers SET status = 2, update_at = NOW() WHERE mailid = ? AND userid = ?"
	_, err = db.MySQLDBGameWeb.Exec(updateQuery, mailID, userID)
	if err != nil {
		return "", err
	}

	return awards, nil
}

// createSystemMail 创建系统邮件
func createSystemMail(tx *sql.Tx, req *models.SendMailRequest, adminID interface{}) (int64, error) {
	// 1. 创建邮件记录
	mailQuery := `
		INSERT INTO mails (type, senderid, title, content, awards, created_at)
		VALUES (?, 0, ?, ?, ?, NOW())
	`
	
	result, err := tx.Exec(mailQuery, req.Type, req.Title, req.Content, req.Awards)
	if err != nil {
		return 0, err
	}

	mailID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 2. 创建系统邮件记录
	systemMailQuery := `
		INSERT INTO mailSystem (type, mailid, startTime, endTime)
		VALUES (?, ?, ?, ?)
	`
	
	_, err = tx.Exec(systemMailQuery, req.Type, mailID, req.StartTime, req.EndTime)
	if err != nil {
		return 0, err
	}

	return mailID, nil
}

// GetAdminMailList 获取管理后台邮件列表
func GetAdminMailList(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	mailType := c.Query("type")
	title := c.Query("title")

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询条件
	var whereConditions []string
	var args []interface{}

	if mailType != "" {
		whereConditions = append(whereConditions, "m.type = ?")
		args = append(args, mailType)
	}

	if title != "" {
		whereConditions = append(whereConditions, "m.title LIKE ?")
		args = append(args, "%"+title+"%")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid 
		%s
	`, whereClause)

	var total int64
	err := db.MySQLDBGameWeb.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Errorf("查询邮件总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询邮件列表
	offset := (page - 1) * pageSize
	listQuery := fmt.Sprintf(`
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       COALESCE(ms.startTime, m.created_at) as startTime,
		       COALESCE(ms.endTime, DATE_ADD(m.created_at, INTERVAL 30 DAY)) as endTime
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid
		%s
		ORDER BY m.id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数
	finalArgs := append(args, pageSize, offset)

	rows, err := db.MySQLDBGameWeb.Query(listQuery, finalArgs...)
	if err != nil {
		log.Errorf("查询邮件列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "查询失败",
		})
		return
	}
	defer rows.Close()

	var mails []models.MailDetailResponse
	for rows.Next() {
		var mail models.MailDetailResponse
		err := rows.Scan(&mail.ID, &mail.Type, &mail.Title, &mail.Content,
			&mail.Awards, &mail.CreatedAt, &mail.StartTime, &mail.EndTime)
		if err != nil {
			log.Errorf("扫描邮件记录失败: %v", err)
			continue
		}
		
		// 判断邮件状态（生效中/已过期）
		now := time.Now()
		if mail.StartTime.After(now) {
			mail.Status = 0 // 未开始
		} else if mail.EndTime.Before(now) {
			mail.Status = 3 // 已过期
		} else {
			mail.Status = 1 // 生效中
		}
		
		mails = append(mails, mail)
	}

	// 获取管理员信息记录日志
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")
	log.Infof("管理员查询邮件列表: 管理员ID=%v, 管理员=%v, 页码=%d, 条件=%s, IP=%s",
		adminId, username, page, whereClause, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "查询成功",
		Data: gin.H{
			"list":     mails,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetAdminMailDetail 获取管理后台邮件详情
func GetAdminMailDetail(c *gin.Context) {
	// 获取邮件ID
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的邮件ID",
		})
		return
	}

	// 查询邮件详情
	query := `
		SELECT m.id, m.type, m.title, m.content, m.awards, m.created_at,
		       COALESCE(ms.startTime, m.created_at) as startTime,
		       COALESCE(ms.endTime, DATE_ADD(m.created_at, INTERVAL 30 DAY)) as endTime,
		       COALESCE(ms.type, m.type) as mailSystemType
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid
		WHERE m.id = ?
	`

	var mail models.MailDetailResponse
	var mailSystemType int
	err = db.MySQLDBGameWeb.QueryRow(query, mailID).Scan(
		&mail.ID, &mail.Type, &mail.Title, &mail.Content, &mail.Awards,
		&mail.CreatedAt, &mail.StartTime, &mail.EndTime, &mailSystemType)

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

	// 判断邮件状态
	now := time.Now()
	if mail.StartTime.After(now) {
		mail.Status = 0 // 未开始
	} else if mail.EndTime.Before(now) {
		mail.Status = 3 // 已过期
	} else {
		mail.Status = 1 // 生效中
	}

	// 查询邮件统计信息（用户领取情况）
	statsQuery := `
		SELECT 
			COUNT(*) as totalUsers,
			COUNT(CASE WHEN status >= 1 THEN 1 END) as readUsers,
			COUNT(CASE WHEN status >= 2 THEN 1 END) as claimedUsers
		FROM mailUsers 
		WHERE mailid = ?
	`

	var stats struct {
		TotalUsers   int64 `json:"totalUsers"`
		ReadUsers    int64 `json:"readUsers"`
		ClaimedUsers int64 `json:"claimedUsers"`
	}

	err = db.MySQLDBGameWeb.QueryRow(statsQuery, mailID).Scan(
		&stats.TotalUsers, &stats.ReadUsers, &stats.ClaimedUsers)
	if err != nil {
		// 统计信息查询失败不影响主数据返回
		log.Warnf("查询邮件统计信息失败: %v", err)
	}

	// 获取管理员信息记录日志
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")
	log.Infof("管理员查询邮件详情: 管理员ID=%v, 管理员=%v, 邮件ID=%d, IP=%s",
		adminId, username, mailID, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "查询成功",
		Data: gin.H{
			"mail":  mail,
			"stats": stats,
		},
	})
}

// UpdateMailStatus 更新邮件状态
func UpdateMailStatus(c *gin.Context) {
	// 获取邮件ID
	mailIDStr := c.Param("id")
	mailID, err := strconv.ParseInt(mailIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的邮件ID",
		})
		return
	}

	// 解析请求参数
	var req struct {
		Action    string    `json:"action" binding:"required"`    // 操作类型: extend, disable
		EndTime   *time.Time `json:"endTime"`                       // 延长截止时间
		StartTime *time.Time `json:"startTime"`                     // 修改开始时间
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("更新邮件状态参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证操作类型
	if req.Action != "extend" && req.Action != "disable" && req.Action != "modify" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "不支持的操作类型",
		})
		return
	}

	// 检查邮件是否存在
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mails WHERE id = ?)"
	err = db.MySQLDBGameWeb.QueryRow(checkQuery, mailID).Scan(&exists)
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

	// 执行操作
	switch req.Action {
	case "extend":
		// 延长邮件有效期
		if req.EndTime == nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "延长操作需要提供新的截止时间",
			})
			return
		}
		
		updateQuery := "UPDATE mailSystem SET endTime = ? WHERE mailid = ?"
		_, err = db.MySQLDBGameWeb.Exec(updateQuery, req.EndTime, mailID)
		
	case "disable":
		// 禁用邮件（设置截止时间为当前时间）
		now := time.Now()
		updateQuery := "UPDATE mailSystem SET endTime = ? WHERE mailid = ?"
		_, err = db.MySQLDBGameWeb.Exec(updateQuery, now, mailID)
		
	case "modify":
		// 修改邮件时间范围
		if req.StartTime == nil || req.EndTime == nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "修改操作需要提供开始和截止时间",
			})
			return
		}
		
		if req.EndTime.Before(*req.StartTime) {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "截止时间不能早于开始时间",
			})
			return
		}
		
		updateQuery := "UPDATE mailSystem SET startTime = ?, endTime = ? WHERE mailid = ?"
		_, err = db.MySQLDBGameWeb.Exec(updateQuery, req.StartTime, req.EndTime, mailID)
	}

	if err != nil {
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
	log.Infof("管理员更新邮件状态: 管理员ID=%v, 管理员=%v, 邮件ID=%d, 操作=%s, IP=%s",
		adminId, username, mailID, req.Action, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "更新成功",
	})
}

// GetMailStats 获取邮件统计
func GetMailStats(c *gin.Context) {
	// 查询邮件统计信息
	var stats struct {
		TotalMails    int64 `json:"totalMails"`    // 总邮件数
		SystemMails   int64 `json:"systemMails"`   // 系统邮件数
		PersonalMails int64 `json:"personalMails"` // 个人邮件数
		ActiveMails   int64 `json:"activeMails"`   // 生效中邮件数
		ExpiredMails  int64 `json:"expiredMails"`  // 已过期邮件数
		PendingMails  int64 `json:"pendingMails"`  // 未开始邮件数
	}

	// 查询总邮件数
	err := db.MySQLDBGameWeb.QueryRow("SELECT COUNT(*) FROM mails").Scan(&stats.TotalMails)
	if err != nil {
		log.Errorf("查询总邮件数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询系统邮件数
	err = db.MySQLDBGameWeb.QueryRow("SELECT COUNT(*) FROM mails WHERE type = 0").Scan(&stats.SystemMails)
	if err != nil {
		log.Errorf("查询系统邮件数失败: %v", err)
		stats.SystemMails = 0
	}

	// 查询个人邮件数
	err = db.MySQLDBGameWeb.QueryRow("SELECT COUNT(*) FROM mails WHERE type = 1").Scan(&stats.PersonalMails)
	if err != nil {
		log.Errorf("查询个人邮件数失败: %v", err)
		stats.PersonalMails = 0
	}

	// 查询生效中邮件数
	now := time.Now()
	activeQuery := `
		SELECT COUNT(*) 
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid 
		WHERE COALESCE(ms.startTime, m.created_at) <= ? 
		  AND COALESCE(ms.endTime, DATE_ADD(m.created_at, INTERVAL 30 DAY)) > ?
	`
	err = db.MySQLDBGameWeb.QueryRow(activeQuery, now, now).Scan(&stats.ActiveMails)
	if err != nil {
		log.Errorf("查询生效中邮件数失败: %v", err)
		stats.ActiveMails = 0
	}

	// 查询已过期邮件数
	expiredQuery := `
		SELECT COUNT(*) 
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid 
		WHERE COALESCE(ms.endTime, DATE_ADD(m.created_at, INTERVAL 30 DAY)) <= ?
	`
	err = db.MySQLDBGameWeb.QueryRow(expiredQuery, now).Scan(&stats.ExpiredMails)
	if err != nil {
		log.Errorf("查询已过期邮件数失败: %v", err)
		stats.ExpiredMails = 0
	}

	// 查询未开始邮件数
	pendingQuery := `
		SELECT COUNT(*) 
		FROM mails m 
		LEFT JOIN mailSystem ms ON m.id = ms.mailid 
		WHERE COALESCE(ms.startTime, m.created_at) > ?
	`
	err = db.MySQLDBGameWeb.QueryRow(pendingQuery, now).Scan(&stats.PendingMails)
	if err != nil {
		log.Errorf("查询未开始邮件数失败: %v", err)
		stats.PendingMails = 0
	}

	// 查询用户邮件统计（最近30天内）
	recentDate := now.AddDate(0, 0, -30)
	userStatsQuery := `
		SELECT 
			COUNT(DISTINCT userid) as totalUsers,
			COUNT(*) as totalUserMails,
			COUNT(CASE WHEN status >= 1 THEN 1 END) as readMails,
			COUNT(CASE WHEN status >= 2 THEN 1 END) as claimedMails
		FROM mailUsers 
		WHERE update_at >= ?
	`

	var userStats struct {
		TotalUsers     int64 `json:"totalUsers"`     // 最近30天内有邮件的用户数
		TotalUserMails int64 `json:"totalUserMails"` // 最近30天内用户邮件总数
		ReadMails      int64 `json:"readMails"`      // 已读邮件数
		ClaimedMails   int64 `json:"claimedMails"`   // 已领取邮件数
	}

	err = db.MySQLDBGameWeb.QueryRow(userStatsQuery, recentDate).Scan(
		&userStats.TotalUsers, &userStats.TotalUserMails, &userStats.ReadMails, &userStats.ClaimedMails)
	if err != nil {
		log.Errorf("查询用户邮件统计失败: %v", err)
		// 设置默认值
		userStats.TotalUsers = 0
		userStats.TotalUserMails = 0
		userStats.ReadMails = 0
		userStats.ClaimedMails = 0
	}

	// 获取管理员信息记录日志
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")
	log.Infof("管理员查询邮件统计: 管理员ID=%v, 管理员=%v, IP=%s",
		adminId, username, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data: gin.H{
			"mailStats": stats,
			"userStats": userStats,
		},
	})
}