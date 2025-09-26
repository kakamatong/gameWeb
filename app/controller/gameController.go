package controller

// 在导入部分添加net/url包
import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"io/ioutil"
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

func getWechatInfo(appid int) (config.WechatInfo, error) {
	for _, info := range config.AppConfig.WechatInfos {
		if info.ID == appid {
			return info, nil
		}
	}
	return config.WechatInfo{}, errors.New("未找到对应的微信配置信息")
}

func ThirdLogin(c *gin.Context) {
	var req struct {
		Appid     int    `json:"appid" binding:"required"`
		LoginType string `json:"loginType" binding:"required"`
		LoginData string `json:"loginData" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request",
		})
		return
	}

	if req.LoginType == "wechatMiniGame" {
		code := req.LoginData
		info, err := getWechatInfo(req.Appid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to login",
			})
			return
		}

		values := url.Values{}
		values.Add("appid", info.AppID)
		values.Add("secret", info.Secret)
		values.Add("js_code", code)
		values.Add("grant_type", "authorization_code")
		baseURL := "https://api.weixin.qq.com/sns/jscode2session"
		fullURL := baseURL + "?" + values.Encode()
		resp, err := http.Get(fullURL)
		if err != nil {
			log.Errorf("Failed to make HTTP request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to make HTTP request",
			})
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read response body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to read response body",
			})
			return
		}

		var wxresp struct {
			SessionKey string `json:"session_key"`
			Unionid    string `json:"unionid"`
			Errmsg     string `json:"errmsg"`
			Errcode    int    `json:"errcode"`
			Openid     string `json:"openid"`
		}

		err = json.Unmarshal(body, &wxresp)
		if err != nil {
			log.Errorf("Failed to unmarshal response body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to unmarshal response body",
			})
			return
		}

		token := md5.Sum([]byte(wxresp.SessionKey))
		// 将 [16]byte 转换为十六进制字符串
		tokenStr := fmt.Sprintf("%x", token)
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Success",
			"data":    map[string]interface{}{"openid": wxresp.Openid, "token": tokenStr},
		})

	}
}
