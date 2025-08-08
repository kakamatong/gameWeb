package controller

import (
	"encoding/json"
	"gameWeb/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 定义服务节点结构体
type ServiceNode struct {
	Addr       string `json:"addr"`
	Name       string `json:"name"`
	Cnt        int    `json:"cnt"`
	ClientAddr string `json:"clientAddr,omitempty"` // omitempty表示如果为空则不序列化
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
		logrus.Errorf("Failed to get clusterConfig from Redis: %v", err)
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
		logrus.Errorf("Failed to parse clusterConfig: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to parse cluster configuration",
			"error":   err.Error(),
		})
		return
	}

	// 返回解析后的数据
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": clusterConfig,
	})
}
