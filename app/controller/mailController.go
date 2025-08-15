package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 邮件列表响应结构体
type MailListResponse struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	MailID int64  `json:"mailid"`
	Status int    `json:"status"`
	Time   string `json:"time"`
}

// 邮件详情响应结构体
// 修改MailDetailResponse结构体定义
type MailDetailResponse struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Awards  string `json:"awards"`
	Status  int    `json:"status"`
	Time    string `json:"time"`
}

// 通用请求结构体
// 获取邮件列表请求
type GetMailListRequest struct {
	UserID int64 `json:"userid" binding:"required"`
}

// 获取邮件详情请求
type GetMailDetailRequest struct {
	UserID int64 `json:"userid" binding:"required"`
}

// 标记邮件已读请求
type MarkMailAsReadRequest struct {
	UserID int64 `json:"userid" binding:"required"`
}

// 领取邮件奖励请求
type GetMailAwardRequest struct {
	UserID int64 `json:"userid" binding:"required"`
}

// 获取邮件列表
func GetMailList(c *gin.Context) {
	var req GetMailListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// 先从系统邮件表拉取未过期的邮件
	syncSystemMails(req.UserID)

	// 查询用户邮件列表
	query := `SELECT mu.id, m.title, mu.mailid, mu.status, mu.update_at 
			FROM mailUsers mu 
			JOIN mails m ON mu.mailid = m.id 
			WHERE mu.userid = ? AND mu.status != 3 
			ORDER BY mu.update_at DESC`

	rows, err := db.MySQLDBGameWeb.Query(query, req.UserID)
	if err != nil {
		log.Errorf("Failed to query mail list for user %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get mail list",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var mailList []MailListResponse
	for rows.Next() {
		var mail MailListResponse
		var updateTime time.Time
		if err := rows.Scan(&mail.ID, &mail.Title, &mail.MailID, &mail.Status, &updateTime); err != nil {
			log.Errorf("Failed to scan mail row: %v", err)
			continue
		}
		mail.Time = updateTime.Format("2006-01-02 15:04:05")
		mailList = append(mailList, mail)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": mailList,
	})
}

// 获取邮件详情
func GetMailDetail(c *gin.Context) {
	var req GetMailDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	mailID := c.Param("id")
	log.Infof("GetMailDetail mailID: %s", mailID)
	if mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "mail id is required",
		})
		return
	}
	mailIDInt, err1 := strconv.ParseInt(mailID, 10, 64)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid mail id",
		})
		return
	}

	log.Infof("GetMailDetail mailIDInt: %d", mailIDInt)
	// 查询邮件详情
	query := `SELECT mu.id, m.title, m.content, m.awards, mu.status, mu.update_at 
			FROM mailUsers mu 
			JOIN mails m ON mu.mailid = m.id 
			WHERE mu.userid = ? AND mu.mailid = ?`

	var mailDetail MailDetailResponse
	var updateTime time.Time
	log.Infof("GetMailDetail query: %d %d", req.UserID, mailIDInt)
	err := db.MySQLDBGameWeb.QueryRow(query, req.UserID, mailIDInt).Scan(
		&mailDetail.ID, &mailDetail.Title, &mailDetail.Content,
		&mailDetail.Awards, &mailDetail.Status, &updateTime,
	)
	log.Infof("GetMailDetail err: %v", err)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found",
			})
		} else {
			log.Errorf("Failed to query mail detail for user %d, mail %s: %v", req.UserID, mailID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to get mail detail",
				"error":   err.Error(),
			})
		}
		return
	}

	mailDetail.Time = updateTime.Format("2006-01-02 15:04:05")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    mailDetail,
	})
}

// 标记邮件为已读
func MarkMailAsRead(c *gin.Context) {
	var req MarkMailAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	mailID := c.Param("id")
	if mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "mail id is required",
		})
		return
	}

	// 更新邮件状态
	query := `UPDATE mailUsers 
			SET status = 1, update_at = CURRENT_TIMESTAMP 
			WHERE userid = ? AND mailid = ? AND status = 0`

	result, err := db.MySQLDBGameWeb.Exec(query, req.UserID, mailID)
	if err != nil {
		log.Errorf("Failed to mark mail as read for user %d, mail %s: %v", req.UserID, mailID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to mark mail as read",
			"error":   err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Mail already read or not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Mail marked as read successfully",
	})
}

// 领取邮件奖励
func GetMailAward(c *gin.Context) {
	// 使用 Gin 上下文的 Request Context
	ctx := c.Request.Context()

	userid, b := c.Get("userid")
	if !b {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userid is required",
		})
		return
	}

	mailID := c.Param("id")
	if mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "mail id is required",
		})
		return
	}

	// 生成锁键，使用用户ID和邮件ID确保唯一性
	userIDInt, ok := userid.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Invalid userid type",
		})
		return
	}
	lockKey := fmt.Sprintf("mail_award_lock:%d:%s", userIDInt, mailID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	lockExpire := 5 * time.Second

	// 尝试获取Redis锁 - 使用 Gin 上下文
	success, err := db.RedisClient.SetNX(ctx, lockKey, lockValue, lockExpire).Result()
	if err != nil {
		log.Errorf("Failed to acquire Redis lock: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "System busy, please try again later",
			"error":   err.Error(),
		})
		return
	}

	if !success {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    429,
			"message": "Operation too frequent, please try again later",
		})
		return
	}

	// 确保函数结束时释放锁
	defer func() {
		// 使用Lua脚本确保只有持有锁的客户端才能释放锁
		script := `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
		// 使用 Gin 上下文
		db.RedisClient.Eval(ctx, script, []string{lockKey}, lockValue)
	}()

	// 开始事务
	tx, err := db.MySQLDBGameWeb.Begin()
	if err != nil {
		log.Errorf("Failed to start transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get award",
			"error":   err.Error(),
		})
		return
	}

	// 查询邮件奖励
	var awards string
	var mailStatus int
	query := `SELECT m.awards, mu.status 
            FROM mailUsers mu 
            JOIN mails m ON mu.mailid = m.id 
            WHERE mu.userid = ? AND mu.mailid = ? FOR UPDATE`
	err = tx.QueryRow(query, userid, mailID).Scan(&awards, &mailStatus)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found",
			})
		} else {
			log.Errorf("Failed to query mail award for user %d, mail %s: %v", userid, mailID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to get award",
				"error":   err.Error(),
			})
		}
		return
	}

	if awards == "" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Mail awards is empty",
		})
		return
	}

	// 检查邮件状态
	if mailStatus == 2 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Mail already got",
		})
		return
	}

	// 解析奖励数据
	var awardData struct {
		Props []struct {
			Id  int `json:"id"`
			Cnt int `json:"cnt"`
		} `json:"props"`
	}

	if err = json.Unmarshal([]byte(awards), &awardData); err != nil {
		tx.Rollback()
		log.Errorf("Failed to parse awards data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to parse awards",
			"error":   err.Error(),
		})
		return
	}

	// 使用主数据库的事务
	mainTx, err := db.MySQLDB.Begin()
	if err != nil {
		tx.Rollback()
		log.Errorf("Failed to start main database transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to start transaction",
			"error":   err.Error(),
		})
		return
	}

	// 发放每个奖励
	for _, prop := range awardData.Props {
		// 检查是否已有记录
		var currentCount int64
		checkQuery := `SELECT richNums FROM userRiches WHERE userid = ? AND richType = ?`
		err = mainTx.QueryRow(checkQuery, userIDInt, prop.Id).Scan(&currentCount)

		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			mainTx.Rollback()
			log.Errorf("Failed to check user riches: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to check user riches",
				"error":   err.Error(),
			})
			return
		}

		// 更新或插入记录
		if err == sql.ErrNoRows {
			// 插入新记录
			insertQuery := `INSERT INTO userRiches (userid, richType, richNums) VALUES (?, ?, ?)`
			_, err = mainTx.Exec(insertQuery, userIDInt, prop.Id, prop.Cnt)
			if err != nil {
				tx.Rollback()
				mainTx.Rollback()
				log.Errorf("Failed to insert user riches: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to insert user riches",
					"error":   err.Error(),
				})
				return
			}
		} else {
			// 更新现有记录
			updateQuery := `UPDATE userRiches SET richNums = richNums + ? WHERE userid = ? AND richType = ?`
			_, err = mainTx.Exec(updateQuery, prop.Cnt, userIDInt, prop.Id)
			if err != nil {
				tx.Rollback()
				mainTx.Rollback()
				log.Errorf("Failed to update user riches: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to update user riches",
					"error":   err.Error(),
				})
				return
			}
		}
	}

	// 提交主数据库事务
	if err = mainTx.Commit(); err != nil {
		tx.Rollback()
		log.Errorf("Failed to commit main database transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
		return
	}

	// 更新邮件状态为已领取
	updateQuery := `UPDATE mailUsers 
            SET status = 2, update_at = CURRENT_TIMESTAMP 
            WHERE userid = ? AND mailid = ?`
	_, err = tx.Exec(updateQuery, userid, mailID)

	if err != nil {
		tx.Rollback()
		log.Errorf("Failed to update mail status for user %d, mail %s: %v", userid, mailID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get award",
			"error":   err.Error(),
		})
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Errorf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get award",
			"error":   err.Error(),
		})
		return
	}

	// 通知游戏服务插入通知
	// 1. 准备奖励数据
	var message struct {
		RichTypes []int `json:"richTypes"`
		RichNums  []int `json:"richNums"`
	}

	for _, prop := range awardData.Props {
		message.RichTypes = append(message.RichTypes, prop.Id)
		message.RichNums = append(message.RichNums, prop.Cnt)
	}

	// 2. 准备请求数据
	reqData := struct {
		UserID  int64       `json:"userid"`
		Message interface{} `json:"awardMessage"`
	}{
		UserID:  userIDInt,
		Message: message,
	}

	// 3. 转换为JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Errorf("Failed to marshal request data: %v", err)
		// 这里不返回错误，因为奖励已经发放成功
	} else {
		// 4. 发送HTTP POST请求
		gameServerConfig := config.AppConfig.GameServer
		url := fmt.Sprintf("http://%s:%s/awardnotice", gameServerConfig.Host, gameServerConfig.Port)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Errorf("Failed to send award notice to game server: %v", err)
		} else {
			defer resp.Body.Close()

			// 5. 解析响应
			var respData struct {
				NoticeID int64 `json:"noticeid"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
				log.Errorf("Failed to decode game server response: %v", err)
			} else {
				log.Infof("Award notice sent successfully, noticeid: %d", respData.NoticeID)
				// 6. 将noticeid返回给客户端
				c.JSON(http.StatusOK, gin.H{
					"code":     200,
					"message":  "Award got successfully",
					"awards":   awards,
					"noticeid": respData.NoticeID,
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Award got successfully",
		"awards":  awards,
	})
}

// 同步系统邮件到用户邮件表
func syncSystemMails(userID int64) {
	// 查询未过期的系统邮件
	now := time.Now().Format("2006-01-02 15:04:05")
	query := `SELECT ms.mailid, ms.startTime, ms.endTime 
			FROM mailSystem ms 
			WHERE ms.endTime > ? 
			AND NOT EXISTS (
				SELECT 1 FROM mailUsers mu 
				WHERE mu.userid = ? AND mu.mailid = ms.mailid
			)`

	rows, err := db.MySQLDBGameWeb.Query(query, now, userID)
	if err != nil {
		log.Errorf("Failed to query system mails: %v", err)
		return
	}
	defer rows.Close()

	// 插入用户邮件
	for rows.Next() {
		var mailID int64
		var startTime, endTime time.Time
		if err := rows.Scan(&mailID, &startTime, &endTime); err != nil {
			log.Errorf("Failed to scan system mail row: %v", err)
			continue
		}

		insertQuery := `INSERT INTO mailUsers (userid, mailid, status, startTime, endTime) 
			VALUES (?, ?, 0, ?, ?)`
		_, err := db.MySQLDBGameWeb.Exec(insertQuery, userID, mailID, startTime, endTime)
		if err != nil {
			log.Errorf("Failed to insert mail for user %d, mail %d: %v", userID, mailID, err)
		}
	}
}
