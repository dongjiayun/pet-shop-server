package controllers

import (
	"github.com/dongjiayun/pet-shop-server/config"
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

func GetDict(c *gin.Context) {
	var key = c.Param("key")
	if key == "breed" {
		var dataString = config.BreedDict
		var data config.Categories
		// 处理 JSON 解析错误
		if err := json.Unmarshal([]byte(dataString), &data); err != nil {
			c.JSON(500, models.Result{Code: 500, Message: "Failed to parse breed dictionary: " + err.Error(), Data: nil})
			return
		}
		c.JSON(200, models.Result{Code: 200, Message: "success", Data: &data})
	} else {
		// 错误请求使用正确的 HTTP 状态码
		c.JSON(400, models.Result{Code: 400, Message: "Invalid dictionary key", Data: nil})
	}
}
