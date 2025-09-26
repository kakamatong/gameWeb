package controller

// 在导入部分添加net/url包
import (
	"encoding/json"
	"gameWeb/db"
	"gameWeb/log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// 定义服务节点结构体
type ServiceNode struct {
	Addr       string `json:"addr"`
	Name       string `json:"name"`
	Cnt        int    `json:"cnt"`
	ClientAddr string `json:"clientAddr,omitempty"`
	Hide       bool   `json:"hide,omitempty"`
}

// 定义ClusterConfig结构体
type ClusterConfig struct {
	List struct {
		Match    []ServiceNode `json:"match"`
		Robot    []ServiceNode `json:"robot"`
		Game     []ServiceNode `json:"game"`
		Login    []ServiceNode `json:"login"`
		User     []ServiceNode `json:"user"`
		Gate     []ServiceNode `json:"gate"`
		Activity []ServiceNode `json:"activity"`
		Auth     []ServiceNode `json:"auth"`
	} `json:"list"`
	Ver int `json:"ver"`
}

// GetAuthGameList 获取游戏列表
func GetAuthGameList(c *gin.Context) {
	// 从Redis获取clusterConfig
	clusterConfigStr, err := db.GetRedis("clusterConfig")
	if err != nil {
		log.Errorf("Failed to get clusterConfig from Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get cluster configuration",
			"error":   err.Error(),
		})
		return
	}

	// 如果获取到空值
	if clusterConfigStr == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Cluster configuration not found",
		})
		return
	}

	// 解析JSON
	var clusterConfig ClusterConfig
	if err := json.Unmarshal([]byte(clusterConfigStr), &clusterConfig); err != nil {
		log.Errorf("Failed to parse clusterConfig: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to parse cluster configuration",
			"error":   err.Error(),
		})
		return
	}

	// 将gate和game根据类型分开存储
	result := make(map[string]map[string]string)
	result["gate"] = make(map[string]string)
	result["game"] = make(map[string]string)
	result["login"] = make(map[string]string)

	// 在处理gate和game数据时添加urlencode处理
	// 处理gate数据
	for _, gate := range clusterConfig.List.Gate {
		if gate.Hide {
			continue
		}
		encodedAddr := url.QueryEscape(gate.ClientAddr)
		result["gate"][gate.Name] = encodedAddr
	}

	// 处理game数据
	for _, game := range clusterConfig.List.Game {
		encodedAddr := url.QueryEscape(game.ClientAddr)
		result["game"][game.Name] = encodedAddr
	}

	for _, login := range clusterConfig.List.Login {
		if login.Hide {
			continue
		}
		encodedAddr := url.QueryEscape(login.ClientAddr)
		result["login"][login.Name] = encodedAddr
	}

	// 返回整理后的数据
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

func ThirdLogin(c *gin.Context){
	var req struct {
		appid int64 `json:"appid" binding:"required"`
		loginType string `json:"loginType" binding:"required"`
		loginData string `json:"loginData" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request",
		})
		return
	}

	if(req.loginType == "wechatMiniGame"){
		code := req.loginData

	}
}
