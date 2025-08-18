package router

import (
	"github.com/dongjiayun/pet-shop-server/controllers"
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func checkTokenMiddleware(c *gin.Context) {
	checkToken, _ := controllers.CheckToken(c)
	if checkToken == nil {
		c.JSON(403, models.Result{Code: 10001, Message: "token is invalid"})
		c.Abort()
		return
	}
	c.Set("cid", checkToken.Cid)
}

func checkAdminMiddleware(c *gin.Context) {
	isAdmin := controllers.CheckIsAdmin(c)
	if !isAdmin {
		c.JSON(403, models.Result{Code: 10001, Message: "权限不足"})
		c.Abort()
		return
	}
}

func checkSuperAdminMiddleware(c *gin.Context) {
	isSuperAdmin := controllers.CheckIsSuperAdmin(c)
	if !isSuperAdmin {
		c.JSON(403, models.Result{Code: 10001, Message: "权限不足"})
		c.Abort()
		return
	}
}

func checkIsEmployeeMiddleware(c *gin.Context) {
	isEmployee := controllers.CheckIsEmployee(c)
	if !isEmployee {
		c.JSON(403, models.Result{Code: 10001, Message: "权限不足"})
		c.Abort()
		return
	}
}

func getAuthApi(router *gin.Engine) {
	r := router.Group("/auth")

	r.POST("/signIn", controllers.SignIn)

	r.POST("/sendEmailOtp", controllers.SendEmailOtp)

	r.POST("/findbackPassword", controllers.FindbackPassword)

	r.Use(checkTokenMiddleware).POST("/signOut", controllers.SignOut)

	r.POST("/refreshToken", controllers.RefreshToken)

	r.POST("/resetPassword", controllers.ResetPassword)
}

func getUserApi(router *gin.Engine) {
	r := router.Group("/user")

	r.Use(checkTokenMiddleware)
	r.Use(checkIsEmployeeMiddleware)

	r.GET(":cid", controllers.GetUser)

	r.PUT("", controllers.UpdateUser)

	r.DELETE(":cid", controllers.DeleteUser)

	r.DELETE("/delete/:cid", controllers.HardDeleteUser)

	r.POST("", controllers.CreateUser)

	r.POST("get", controllers.GetUsers)

	r.Use(checkSuperAdminMiddleware)
	r.PUT("permission", controllers.SetPermission)
}

func getPetApi(router *gin.Engine) {
	r := router.Group("/pet")

	r.Use(checkTokenMiddleware)
	r.Use(checkIsEmployeeMiddleware)

	r.GET(":PetId", controllers.GetPet)

	r.POST("list", controllers.GetPets)

	r.POST("", controllers.CreatePet)

	r.PUT(":PetId", controllers.UpdatePet)

	r.DELETE(":PetId", controllers.DeletePet)
}

func getDictApi(router *gin.Engine) {
	r := router.Group("/dict")

	r.Use(checkTokenMiddleware)
	r.Use(checkIsEmployeeMiddleware)

	r.GET(":key", controllers.GetDict)
}

func setCros(router *gin.Engine) {
	router.Use(CORSMiddleware())
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	setCros(router)
	getAuthApi(router)
	getUserApi(router)
	getDictApi(router)
	getPetApi(router)
	return router
}
