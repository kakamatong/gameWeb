package controller

import (
	"database/sql"
	"fmt"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取用户列表
func GetUserList(c *gin.Context) {
	var req models.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Errorf("用户列表参数绑定失败: %v", err)
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
		whereConditions = append(whereConditions, "u.userid = ?")
		args = append(args, req.UserID)
	}

	if req.Keyword != "" {
		whereConditions = append(whereConditions, "u.nickname LIKE ?")
		args = append(args, "%"+req.Keyword+"%")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	total, err := getUserCount(whereClause, args)
	if err != nil {
		log.Errorf("查询用户总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	// 查询用户列表
	users, err := getUserList(whereClause, args, req.Page, req.PageSize)
	if err != nil {
		log.Errorf("查询用户列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	response := models.UserListResponse{
		Total: total,
		Users: users,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// GetUserDetail 获取用户详细信息
func GetUserDetail(c *gin.Context) {
	userIDStr := c.Param("userid")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	user, err := getUserDetailByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Code:    404,
				Message: "用户不存在",
			})
			return
		}
		log.Errorf("查询用户详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    user,
	})
}

// UpdateUser 更新用户信息
func UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userid")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("更新用户参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证用户是否存在
	exists, err := checkUserExists(userID)
	if err != nil {
		log.Errorf("检查用户存在性失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	// 开始事务
	tx, err := db.MySQLDB.Begin()
	if err != nil {
		log.Errorf("开始事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	defer tx.Rollback()

	// 更新用户基础信息
	if req.Nickname != "" {
		if err := updateUserNickname(tx, userID, req.Nickname); err != nil {
			log.Errorf("更新用户昵称失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "更新失败",
			})
			return
		}
	}

	// 更新用户状态
	if req.Status != nil {
		if err := updateUserStatus(tx, userID, *req.Status); err != nil {
			log.Errorf("更新用户状态失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "更新失败",
			})
			return
		}
	}

	// 更新用户财富
	if req.Riches != nil && len(req.Riches) > 0 {
		if err := updateUserRiches(tx, userID, req.Riches); err != nil {
			log.Errorf("更新用户财富失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    500,
				Message: "更新失败",
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Errorf("提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	// 获取管理员信息记录操作日志
	adminId, _ := c.Get("adminId")
	username, _ := c.Get("username")
	log.Infof("管理员更新用户信息: 管理员ID=%v, 管理员=%v, 用户ID=%d, IP=%s", 
		adminId, username, userID, c.ClientIP())

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "更新成功",
	})
}

// 数据库操作函数

// getUserCount 获取用户总数
func getUserCount(whereClause string, args []interface{}) (int64, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(DISTINCT u.userid) 
		FROM userData u
		LEFT JOIN userStatus us ON u.userid = us.userid
		%s
	`, whereClause)

	var count int64
	err := db.MySQLDB.QueryRow(query, args...).Scan(&count)
	return count, err
}

// getUserList 获取用户列表
func getUserList(whereClause string, args []interface{}, page, pageSize int) ([]models.UserInfo, error) {
	offset := (page - 1) * pageSize
	
	query := fmt.Sprintf(`
		SELECT 
			u.userid, u.nickname, u.headurl, u.sex, u.province, u.city, u.ip,
			COALESCE(us.status, 0) as status,
			COALESCE(us.gameid, 0) as gameid,
			COALESCE(us.roomid, 0) as roomid,
			u.create_time, u.update_time
		FROM userData u
		LEFT JOIN userStatus us ON u.userid = us.userid
		%s
		ORDER BY u.userid DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数到args
	finalArgs := append(args, pageSize, offset)
	
	rows, err := db.MySQLDB.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserInfo
	userIDs := []int64{}

	for rows.Next() {
		var user models.UserInfo
		err := rows.Scan(
			&user.UserID, &user.Nickname, &user.HeadURL, &user.Sex,
			&user.Province, &user.City, &user.IP, &user.Status,
			&user.GameID, &user.RoomID, &user.CreateTime, &user.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
		userIDs = append(userIDs, user.UserID)
	}

	// 批量查询用户财富信息
	if len(userIDs) > 0 {
		richesMap, err := getUserRichesBatch(userIDs)
		if err != nil {
			log.Errorf("批量查询用户财富失败: %v", err)
		} else {
			// 将财富信息分配到对应用户
			for i := range users {
				if riches, exists := richesMap[users[i].UserID]; exists {
					users[i].Riches = riches
				} else {
					users[i].Riches = []models.UserRich{}
				}
			}
		}
	}

	return users, nil
}

// getUserDetailByID 根据ID获取用户详细信息
func getUserDetailByID(userID int64) (*models.UserInfo, error) {
	query := `
		SELECT 
			u.userid, u.nickname, u.headurl, u.sex, u.province, u.city, u.ip,
			COALESCE(us.status, 0) as status,
			COALESCE(us.gameid, 0) as gameid,
			COALESCE(us.roomid, 0) as roomid,
			u.create_time, u.update_time
		FROM userData u
		LEFT JOIN userStatus us ON u.userid = us.userid
		WHERE u.userid = ?
	`

	var user models.UserInfo
	err := db.MySQLDB.QueryRow(query, userID).Scan(
		&user.UserID, &user.Nickname, &user.HeadURL, &user.Sex,
		&user.Province, &user.City, &user.IP, &user.Status,
		&user.GameID, &user.RoomID, &user.CreateTime, &user.UpdateTime,
	)

	if err != nil {
		return nil, err
	}

	// 查询用户财富信息
	riches, err := getUserRiches(userID)
	if err != nil {
		log.Errorf("查询用户财富失败: %v", err)
		user.Riches = []models.UserRich{}
	} else {
		user.Riches = riches
	}

	return &user, nil
}

// getUserRiches 获取用户财富信息
func getUserRiches(userID int64) ([]models.UserRich, error) {
	query := "SELECT richType, richNums FROM userRiches WHERE userid = ?"
	rows, err := db.MySQLDB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var riches []models.UserRich
	for rows.Next() {
		var rich models.UserRich
		err := rows.Scan(&rich.RichType, &rich.RichNums)
		if err != nil {
			return nil, err
		}
		riches = append(riches, rich)
	}

	return riches, nil
}

// getUserRichesBatch 批量获取用户财富信息
func getUserRichesBatch(userIDs []int64) (map[int64][]models.UserRich, error) {
	if len(userIDs) == 0 {
		return make(map[int64][]models.UserRich), nil
	}

	// 构建IN子句
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT userid, richType, richNums FROM userRiches WHERE userid IN (%s)", 
		strings.Join(placeholders, ","))

	rows, err := db.MySQLDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	richesMap := make(map[int64][]models.UserRich)
	for rows.Next() {
		var userID int64
		var rich models.UserRich
		err := rows.Scan(&userID, &rich.RichType, &rich.RichNums)
		if err != nil {
			return nil, err
		}
		richesMap[userID] = append(richesMap[userID], rich)
	}

	return richesMap, nil
}

// checkUserExists 检查用户是否存在
func checkUserExists(userID int64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM userData WHERE userid = ?"
	err := db.MySQLDB.QueryRow(query, userID).Scan(&count)
	return count > 0, err
}

// updateUserNickname 更新用户昵称
func updateUserNickname(tx *sql.Tx, userID int64, nickname string) error {
	query := "UPDATE userData SET nickname = ?, update_time = CURRENT_TIMESTAMP WHERE userid = ?"
	_, err := tx.Exec(query, nickname, userID)
	return err
}

// updateUserStatus 更新用户状态
func updateUserStatus(tx *sql.Tx, userID int64, status int8) error {
	// 首先检查userStatus表中是否存在该用户记录
	var count int
	checkQuery := "SELECT COUNT(*) FROM userStatus WHERE userid = ?"
	err := tx.QueryRow(checkQuery, userID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// 更新现有记录
		query := "UPDATE userStatus SET status = ?, update_time = CURRENT_TIMESTAMP WHERE userid = ?"
		_, err = tx.Exec(query, status, userID)
	} else {
		// 插入新记录
		query := "INSERT INTO userStatus (userid, status, gameid, roomid, create_time, update_time) VALUES (?, ?, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"
		_, err = tx.Exec(query, userID, status)
	}
	
	return err
}

// updateUserRiches 更新用户财富
func updateUserRiches(tx *sql.Tx, userID int64, riches []models.UserRich) error {
	for _, rich := range riches {
		// 检查该类型财富是否存在
		var count int
		checkQuery := "SELECT COUNT(*) FROM userRiches WHERE userid = ? AND richType = ?"
		err := tx.QueryRow(checkQuery, userID, rich.RichType).Scan(&count)
		if err != nil {
			return err
		}

		if count > 0 {
			// 更新现有记录
			query := "UPDATE userRiches SET richNums = ? WHERE userid = ? AND richType = ?"
			_, err = tx.Exec(query, rich.RichNums, userID, rich.RichType)
		} else {
			// 插入新记录
			query := "INSERT INTO userRiches (userid, richType, richNums) VALUES (?, ?, ?)"
			_, err = tx.Exec(query, userID, rich.RichType, rich.RichNums)
		}
		
		if err != nil {
			return err
		}
	}
	
	return nil
}