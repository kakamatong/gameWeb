package controller

import (
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

// GetUserAuthLogs 获取用户登录认证日志
func GetUserAuthLogs(c *gin.Context) {
	var req models.LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Errorf("用户登录日志参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建查询条件
	whereConditions := []string{}
	args := []interface{}{}

	if req.UserID > 0 {
		whereConditions = append(whereConditions, "userid = ?")
		args = append(args, req.UserID)
	}

	if !req.StartTime.IsZero() {
		whereConditions = append(whereConditions, "create_time >= ?")
		args = append(args, req.StartTime)
	}

	if !req.EndTime.IsZero() {
		whereConditions = append(whereConditions, "create_time <= ?")
		args = append(args, req.EndTime)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	total, err := getAuthLogCount(whereClause, args)
	if err != nil {
		log.Errorf("查询登录日志总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询日志列表
	logs, err := getAuthLogList(whereClause, args, req.Page, req.PageSize)
	if err != nil {
		log.Errorf("查询登录日志列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	response := models.PaginationResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     logs,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// GetUserGameLogs 获取用户对局结果日志
func GetUserGameLogs(c *gin.Context) {
	var req models.LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Errorf("用户对局日志参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建查询条件
	whereConditions := []string{}
	args := []interface{}{}

	if req.UserID > 0 {
		whereConditions = append(whereConditions, "userid = ?")
		args = append(args, req.UserID)
	}

	if !req.StartTime.IsZero() {
		whereConditions = append(whereConditions, "time >= ?")
		args = append(args, req.StartTime)
	}

	if !req.EndTime.IsZero() {
		whereConditions = append(whereConditions, "time <= ?")
		args = append(args, req.EndTime)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	total, err := getGameLogCount(whereClause, args)
	if err != nil {
		log.Errorf("查询对局日志总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询日志列表
	logs, err := getGameLogList(whereClause, args, req.Page, req.PageSize)
	if err != nil {
		log.Errorf("查询对局日志列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	response := models.PaginationResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     logs,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// GetUserLoginStats 获取用户登录统计信息
func GetUserLoginStats(c *gin.Context) {
	userIDStr := c.Query("userid")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "缺少用户ID参数",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	stats, err := getUserLoginStats(userID)
	if err != nil {
		log.Errorf("获取用户登录统计失败: %v", err)
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

// GetUserGameStats 获取用户对局统计信息
func GetUserGameStats(c *gin.Context) {
	userIDStr := c.Query("userid")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "缺少用户ID参数",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	stats, err := getUserGameStats(userID)
	if err != nil {
		log.Errorf("获取用户对局统计失败: %v", err)
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

// 数据库操作函数 - 认证日志相关

// getAuthLogCount 获取认证日志总数
func getAuthLogCount(whereClause string, args []interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM logAuth %s", whereClause)
	
	var count int64
	err := db.MySQLDBGameLog.QueryRow(query, args...).Scan(&count)
	return count, err
}

// getAuthLogList 获取认证日志列表
func getAuthLogList(whereClause string, args []interface{}, page, pageSize int) ([]models.LogAuth, error) {
	offset := (page - 1) * pageSize
	
	query := fmt.Sprintf(`
		SELECT id, userid, nickname, ip, loginType, status, ext, create_time
		FROM logAuth 
		%s
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数到args
	finalArgs := append(args, pageSize, offset)
	
	rows, err := db.MySQLDBGameLog.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogAuth
	for rows.Next() {
		var logAuth models.LogAuth
		
		err := rows.Scan(
			&logAuth.ID, &logAuth.UserID, &logAuth.Nickname, &logAuth.IP,
			&logAuth.LoginType, &logAuth.Status, &logAuth.Ext, &logAuth.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		
		logs = append(logs, logAuth)
	}

	return logs, nil
}

// getUserLoginStats 获取用户登录统计
func getUserLoginStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询总登录次数
	var totalLogins int64
	query := "SELECT COUNT(*) FROM logAuth WHERE userid = ?"
	err := db.MySQLDBGameLog.QueryRow(query, userID).Scan(&totalLogins)
	if err != nil {
		return nil, err
	}
	stats["totalLogins"] = totalLogins

	// 查询最后登录时间
	var lastLoginTime time.Time
	query = "SELECT MAX(create_time) FROM logAuth WHERE userid = ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID).Scan(&lastLoginTime)
	if err != nil {
		log.Warnf("查询最后登录时间失败: %v", err)
		stats["lastLoginTime"] = nil
	} else {
		stats["lastLoginTime"] = lastLoginTime
	}

	// 查询今日登录次数
	today := time.Now().Format("2006-01-02")
	var todayLogins int64
	query = "SELECT COUNT(*) FROM logAuth WHERE userid = ? AND DATE(create_time) = ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID, today).Scan(&todayLogins)
	if err != nil {
		log.Warnf("查询今日登录次数失败: %v", err)
		stats["todayLogins"] = 0
	} else {
		stats["todayLogins"] = todayLogins
	}

	// 查询本周登录次数
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
	var weekLogins int64
	query = "SELECT COUNT(*) FROM logAuth WHERE userid = ? AND DATE(create_time) >= ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID, weekStart).Scan(&weekLogins)
	if err != nil {
		log.Warnf("查询本周登录次数失败: %v", err)
		stats["weekLogins"] = 0
	} else {
		stats["weekLogins"] = weekLogins
	}

	// 成功登录次数
	var successLogins int64
	query = "SELECT COUNT(*) FROM logAuth WHERE userid = ? AND status = 1"
	err = db.MySQLDBGameLog.QueryRow(query, userID).Scan(&successLogins)
	if err != nil {
		log.Warnf("查询成功登录次数失败: %v", err)
		stats["successLogins"] = 0
	} else {
		stats["successLogins"] = successLogins
	}

	return stats, nil
}

// 数据库操作函数 - 对局日志相关

// getGameLogCount 获取对局日志总数
func getGameLogCount(whereClause string, args []interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM logResult10001 %s", whereClause)
	
	var count int64
	err := db.MySQLDBGameLog.QueryRow(query, args...).Scan(&count)
	return count, err
}

// getGameLogList 获取对局日志列表
func getGameLogList(whereClause string, args []interface{}, page, pageSize int) ([]models.LogResult10001, error) {
	offset := (page - 1) * pageSize
	
	query := fmt.Sprintf(`
		SELECT id, type, userid, gameid, roomid, result, score1, score2, score3, score4, score5, time, ext
		FROM logResult10001 
		%s
		ORDER BY time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数到args
	finalArgs := append(args, pageSize, offset)
	
	rows, err := db.MySQLDBGameLog.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogResult10001
	for rows.Next() {
		var logResult models.LogResult10001
		
		err := rows.Scan(
			&logResult.ID, &logResult.Type, &logResult.UserID, &logResult.GameID, &logResult.RoomID,
			&logResult.Result, &logResult.Score1, &logResult.Score2, &logResult.Score3,
			&logResult.Score4, &logResult.Score5, &logResult.Time, &logResult.Ext,
		)
		if err != nil {
			return nil, err
		}
		
		logs = append(logs, logResult)
	}

	return logs, nil
}

// getUserGameStats 获取用户对局统计
func getUserGameStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询总对局次数
	var totalGames int64
	query := "SELECT COUNT(*) FROM logResult10001 WHERE userid = ?"
	err := db.MySQLDBGameLog.QueryRow(query, userID).Scan(&totalGames)
	if err != nil {
		return nil, err
	}
	stats["totalGames"] = totalGames

	// 查询胜利次数
	var winGames int64
	query = "SELECT COUNT(*) FROM logResult10001 WHERE userid = ? AND result = 1"
	err = db.MySQLDBGameLog.QueryRow(query, userID).Scan(&winGames)
	if err != nil {
		log.Warnf("查询胜利次数失败: %v", err)
		stats["winGames"] = 0
	} else {
		stats["winGames"] = winGames
	}

	// 计算胜率
	if totalGames > 0 {
		winRate := float64(winGames) / float64(totalGames) * 100
		stats["winRate"] = fmt.Sprintf("%.2f%%", winRate)
	} else {
		stats["winRate"] = "0.00%"
	}

	// 查询总赢得财富（score1-5求和）
	var totalScore1, totalScore2, totalScore3, totalScore4, totalScore5 int64
	query = "SELECT COALESCE(SUM(score1), 0), COALESCE(SUM(score2), 0), COALESCE(SUM(score3), 0), COALESCE(SUM(score4), 0), COALESCE(SUM(score5), 0) FROM logResult10001 WHERE userid = ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID).Scan(&totalScore1, &totalScore2, &totalScore3, &totalScore4, &totalScore5)
	if err != nil {
		log.Warnf("查询总赢得财富失败: %v", err)
		stats["totalScore1"] = 0
		stats["totalScore2"] = 0
		stats["totalScore3"] = 0
		stats["totalScore4"] = 0
		stats["totalScore5"] = 0
	} else {
		stats["totalScore1"] = totalScore1
		stats["totalScore2"] = totalScore2
		stats["totalScore3"] = totalScore3
		stats["totalScore4"] = totalScore4
		stats["totalScore5"] = totalScore5
		stats["totalScore"] = totalScore1 + totalScore2 + totalScore3 + totalScore4 + totalScore5
	}

	// 查询最后对局时间
	var lastGameTime time.Time
	query = "SELECT MAX(time) FROM logResult10001 WHERE userid = ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID).Scan(&lastGameTime)
	if err != nil {
		log.Warnf("查询最后对局时间失败: %v", err)
		stats["lastGameTime"] = nil
	} else {
		stats["lastGameTime"] = lastGameTime
	}

	// 查询今日对局次数
	today := time.Now().Format("2006-01-02")
	var todayGames int64
	query = "SELECT COUNT(*) FROM logResult10001 WHERE userid = ? AND DATE(time) = ?"
	err = db.MySQLDBGameLog.QueryRow(query, userID, today).Scan(&todayGames)
	if err != nil {
		log.Warnf("查询今日对局次数失败: %v", err)
		stats["todayGames"] = 0
	} else {
		stats["todayGames"] = todayGames
	}

	return stats, nil
}