package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminAccount 管理员账户模型
type AdminAccount struct {
	ID              uint64    `json:"id" db:"id"`
	Username        string    `json:"username" db:"username"`
	PasswordHash    string    `json:"-" db:"passwordHash"`
	Email           string    `json:"email" db:"email"`
	Mobile          string    `json:"mobile" db:"mobile"`
	Status          int8      `json:"status" db:"status"`
	IsSuperAdmin    int8      `json:"isSuperAdmin" db:"isSuperAdmin"`
	RealName        string    `json:"realName" db:"realName"`
	Avatar          string    `json:"avatar" db:"avatar"`
	DepartmentID    *int      `json:"departmentId" db:"departmentId"`
	Note            string    `json:"note" db:"note"`
	LastLoginIP     string    `json:"lastLoginIp" db:"lastLoginIp"`
	LastLoginTime   time.Time `json:"lastLoginTime" db:"lastLoginTime"`
	CreatedBy       *uint64   `json:"createdBy" db:"createdBy"`
	UpdatedBy       *uint64   `json:"updatedBy" db:"updatedBy"`
	CreatedTime     time.Time `json:"createdTime" db:"createdTime"`
	UpdatedTime     time.Time `json:"updatedTime" db:"updatedTime"`
}

// AdminJWTClaims 管理员JWT声明结构体
type AdminJWTClaims struct {
	AdminID      uint64 `json:"adminId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	RealName     string `json:"realName"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
	DepartmentID *int   `json:"departmentId,omitempty"`
	LoginIP      string `json:"loginIp"`
	jwt.RegisteredClaims
}

// UserData 用户数据模型
type UserData struct {
	UserID     int64     `json:"userid" db:"userid"`
	Nickname   string    `json:"nickname" db:"nickname"`
	HeadURL    string    `json:"headurl" db:"headurl"`
	Sex        int8      `json:"sex" db:"sex"`
	Province   string    `json:"province" db:"province"`
	City       string    `json:"city" db:"city"`
	IP         string    `json:"ip" db:"ip"`
	Ext        string    `json:"ext" db:"ext"`
	CreateTime time.Time `json:"createTime" db:"create_time"`
	UpdateTime time.Time `json:"updateTime" db:"update_time"`
}

// UserRiches 用户财富模型
type UserRiches struct {
	UserID   int64 `json:"userid" db:"userid"`
	RichType int8  `json:"richType" db:"richType"`
	RichNums int64 `json:"richNums" db:"richNums"`
}

// UserStatus 用户状态模型
type UserStatus struct {
	UserID     int64     `json:"userid" db:"userid"`
	Status     int8      `json:"status" db:"status"`
	GameID     int64     `json:"gameid" db:"gameid"`
	RoomID     int64     `json:"roomid" db:"roomid"`
	CreateTime time.Time `json:"createTime" db:"create_time"`
	UpdateTime time.Time `json:"updateTime" db:"update_time"`
}

// UserInfo 用户信息聚合模型（用于API响应）
type UserInfo struct {
	UserID     int64       `json:"userid"`
	Nickname   string      `json:"nickname"`
	HeadURL    string      `json:"headurl"`
	Sex        int8        `json:"sex"`
	Province   string      `json:"province"`
	City       string      `json:"city"`
	IP         string      `json:"ip"`
	Status     int8        `json:"status"`
	GameID     int64       `json:"gameid"`
	RoomID     int64       `json:"roomid"`
	Riches     []UserRich  `json:"riches"`
	CreateTime time.Time   `json:"createTime"`
	UpdateTime time.Time   `json:"updateTime"`
}

// UserRich 用户财富简化模型（用于API响应）
type UserRich struct {
	RichType int8  `json:"richType"`
	RichNums int64 `json:"richNums"`
}

// LogAuth 登录认证日志模型
type LogAuth struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"userid" db:"userid"`
	Nickname   string    `json:"nickname" db:"nickname"`
	IP         string    `json:"ip" db:"ip"`
	LoginType  string    `json:"loginType" db:"loginType"`
	Status     int8      `json:"status" db:"status"`
	Ext        string    `json:"ext" db:"ext"`
	CreateTime time.Time `json:"createTime" db:"create_time"`
}

// LogResult10001 对局结果日志模型
type LogResult10001 struct {
	ID         int64     `json:"id" db:"id"`
	Type       int8      `json:"type" db:"type"`
	UserID     int64     `json:"userid" db:"userid"`
	GameID     int64     `json:"gameid" db:"gameid"`
	RoomID     int64     `json:"roomid" db:"roomid"`
	Result     int8      `json:"result" db:"result"`
	Score1     int64     `json:"score1" db:"score1"`
	Score2     int64     `json:"score2" db:"score2"`
	Score3     int64     `json:"score3" db:"score3"`
	Score4     int64     `json:"score4" db:"score4"`
	Score5     int64     `json:"score5" db:"score5"`
	Time       time.Time `json:"time" db:"time"`
	Ext        string    `json:"ext" db:"ext"`
}

// Mails 邮件模型
type Mails struct {
	ID        int64     `json:"id" db:"id"`
	Type      int8      `json:"type" db:"type"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Awards    string    `json:"awards" db:"awards"`
	StartTime time.Time `json:"startTime" db:"startTime"`
	EndTime   time.Time `json:"endTime" db:"endTime"`
	Status    int8      `json:"status" db:"status"`
}

// MailSystem 邮件系统模型
type MailSystem struct {
	ID      int64 `json:"id" db:"id"`
	MailID  int64 `json:"mailId" db:"mailid"`
	UserID  int64 `json:"userid" db:"userid"`
	IsGlobal int8 `json:"isGlobal" db:"isGlobal"`
}

// MailUsers 用户邮件模型
type MailUsers struct {
	ID         int64     `json:"id" db:"id"`
	MailID     int64     `json:"mailId" db:"mailid"`
	UserID     int64     `json:"userid" db:"userid"`
	IsRead     int8      `json:"isRead" db:"isRead"`
	IsReceived int8      `json:"isReceived" db:"isReceived"`
	ReadTime   time.Time `json:"readTime" db:"readTime"`
	ReceiveTime time.Time `json:"receiveTime" db:"receiveTime"`
	CreateTime time.Time `json:"createTime" db:"create_time"`
}

// API请求和响应模型

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token     string       `json:"token"`
	AdminInfo *AdminInfo   `json:"adminInfo"`
}

// AdminInfo 管理员信息响应
type AdminInfo struct {
	ID            uint64    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Mobile        string    `json:"mobile"`
	RealName      string    `json:"realName"`
	Avatar        string    `json:"avatar"`
	DepartmentID  *int      `json:"departmentId,omitempty"`
	Note          string    `json:"note"`
	Status        int8      `json:"status"`
	IsSuperAdmin  bool      `json:"isSuperAdmin"`
	LastLoginIP   string    `json:"lastLoginIp"`
	LastLoginTime time.Time `json:"lastLoginTime"`
	CreatedBy     *uint64   `json:"createdBy,omitempty"`
	UpdatedBy     *uint64   `json:"updatedBy,omitempty"`
	CreatedTime   time.Time `json:"createdTime"`
	UpdatedTime   time.Time `json:"updatedTime"`
}

// UserListRequest 用户列表查询请求
type UserListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"pageSize,default=20" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	UserID   int64  `form:"userid"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int64      `json:"total"`
	Users []UserInfo `json:"users"`
}

// UserUpdateRequest 用户信息修改请求
type UserUpdateRequest struct {
	Nickname string     `json:"nickname"`
	Status   *int8      `json:"status"`
	Riches   []UserRich `json:"riches"`
}

// LogQueryRequest 日志查询请求
type LogQueryRequest struct {
	UserID    int64     `form:"userid"`
	StartTime time.Time `form:"startTime"`
	EndTime   time.Time `form:"endTime"`
	Page      int       `form:"page,default=1" binding:"min=1"`
	PageSize  int       `form:"pageSize,default=20" binding:"min=1,max=100"`
}

// SendMailRequest 发送邮件请求
type SendMailRequest struct {
	Type        int8      `json:"type" binding:"required"`
	Title       string    `json:"title" binding:"required,min=1,max=100"`
	Content     string    `json:"content" binding:"required,min=1,max=1000"`
	Awards      string    `json:"awards"`
	StartTime   time.Time `json:"startTime" binding:"required"`
	EndTime     time.Time `json:"endTime" binding:"required"`
	TargetUsers []int64   `json:"targetUsers"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Data     interface{} `json:"data"`
}