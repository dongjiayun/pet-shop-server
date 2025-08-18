package main

import (
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/dongjiayun/pet-shop-server/router"
	"github.com/dongjiayun/pet-shop-server/utils"
)

func init() {
	models.InitRedis()
	utils.InitValidator()
}

func main() {
	r := router.GetRouter()
	r.Run("0.0.0.0:2088")
}
