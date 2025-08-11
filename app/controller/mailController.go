package controller

import (
	"database/sql"
	"gameWeb/db"
	"gameWeb/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 邮件列表响应结构体
type MailListResponse struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Time   string `json:"time"`
}

// 邮件详情响应结构体
type MailDetailResponse struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Awards  string `json:"awards"`
	Status  int    `json:"status"`
	Time    string `json:"time"`
}

// 获取邮件列表
func GetMailList(c *gin.Context) {
	userID := c.PostForm("userid")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userid is required",
		})
		return
	}

	// 先从系统邮件表拉取未过期的邮件
	syncSystemMails(userID)

	// 查询用户邮件列表
	query := `SELECT mu.id, m.title, mu.status, mu.update_at 
			FROM mailUsers mu 
			JOIN mails m ON mu.mailid = m.id 
			WHERE mu.userid = ? AND mu.status != 3 
			ORDER BY mu.update_at DESC`

	rows, err := db.MySQLDB.Query(query, userID)
	if err != nil {
		log.Logger.Errorf("Failed to query mail list for user %s: %v", userID, err)
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
		if err := rows.Scan(&mail.ID, &mail.Title, &mail.Status, &updateTime); err != nil {
			log.Logger.Errorf("Failed to scan mail row: %v", err)
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
	userID := c.PostForm("userid")
	mailID := c.Param("id")

	if userID == "" || mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userid and mail id are required",
		})
		return
	}

	// 查询邮件详情
	query := `SELECT mu.id, m.title, m.content, m.awards, mu.status, mu.update_at 
			FROM mailUsers mu 
			JOIN mails m ON mu.mailid = m.id 
			WHERE mu.userid = ? AND mu.id = ?`

	var mailDetail MailDetailResponse
	var updateTime time.Time
	err := db.MySQLDB.QueryRow(query, userID, mailID).Scan(
		&mailDetail.ID, &mailDetail.Title, &mailDetail.Content,
		&mailDetail.Awards, &mailDetail.Status, &updateTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found",
			})
		} else {
			log.Logger.Errorf("Failed to query mail detail for user %s, mail %s: %v", userID, mailID, err)
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
		"code": 200,
		"data": mailDetail,
	})
}

// 标记邮件为已读
func MarkMailAsRead(c *gin.Context) {
	userID := c.PostForm("userid")
	mailID := c.Param("id")

	if userID == "" || mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userid and mail id are required",
		})
		return
	}

	// 更新邮件状态
	query := `UPDATE mailUsers 
			SET status = 1, update_at = CURRENT_TIMESTAMP 
			WHERE userid = ? AND id = ? AND status = 0`

	result, err := db.MySQLDB.Exec(query, userID, mailID)
	if err != nil {
		log.Logger.Errorf("Failed to mark mail as read for user %s, mail %s: %v", userID, mailID, err)
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
	userID := c.PostForm("userid")
	mailID := c.Param("id")

	if userID == "" || mailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userid and mail id are required",
		})
		return
	}

	// 开始事务
	tx, err := db.MySQLDB.Begin()
	if err != nil {
		log.Logger.Errorf("Failed to start transaction: %v", err)
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
			WHERE mu.userid = ? AND mu.id = ? FOR UPDATE`
	err = tx.QueryRow(query, userID, mailID).Scan(&awards, &mailStatus)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Mail not found",
			})
		} else {
			log.Logger.Errorf("Failed to query mail award for user %s, mail %s: %v", userID, mailID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to get award",
				"error":   err.Error(),
			})
		}
		return
	}

	// 检查邮件状态
	if mailStatus != 1 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Mail must be read before getting award",
		})
		return
	}

	// TODO: 发放奖励逻辑
	// 这里应该根据awards字段的内容，给用户发放相应的奖励
	// 例如：解析JSON格式的awards，然后更新用户的游戏数据

	// 更新邮件状态为已领取
	updateQuery := `UPDATE mailUsers 
			SET status = 2, update_at = CURRENT_TIMESTAMP 
			WHERE userid = ? AND id = ?`
	_, err = tx.Exec(updateQuery, userID, mailID)

	if err != nil {
		tx.Rollback()
		log.Logger.Errorf("Failed to update mail status for user %s, mail %s: %v", userID, mailID, err)
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
		log.Logger.Errorf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get award",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Award got successfully",
		"awards":  awards,
	})
}

// 同步系统邮件到用户邮件表
func syncSystemMails(userID string) {
	// 查询未过期的系统邮件
	now := time.Now().Format("2006-01-02 15:04:05")
	query := `SELECT ms.mailid, ms.startTime, ms.endTime 
			FROM mailSystem ms 
			WHERE ms.endTime > ? 
			AND NOT EXISTS (
				SELECT 1 FROM mailUsers mu 
				WHERE mu.userid = ? AND mu.mailid = ms.mailid
			)`

	rows, err := db.MySQLDB.Query(query, now, userID)
	if err != nil {
		log.Logger.Errorf("Failed to query system mails: %v", err)
		return
	}
	defer rows.Close()

	// 插入用户邮件
	for rows.Next() {
		var mailID int64
		var startTime, endTime time.Time
		if err := rows.Scan(&mailID, &startTime, &endTime); err != nil {
			log.Logger.Errorf("Failed to scan system mail row: %v", err)
			continue
		}

		insertQuery := `INSERT INTO mailUsers (userid, mailid, status, startTime, endTime) 
			VALUES (?, ?, 0, ?, ?)`
		_, err := db.MySQLDB.Exec(insertQuery, userID, mailID, startTime, endTime)
		if err != nil {
			log.Logger.Errorf("Failed to insert mail for user %s, mail %d: %v", userID, mailID, err)
		}
	}
}
