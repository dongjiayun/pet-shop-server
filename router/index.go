package router

import (
	"net/http"

	"github.com/dongjiayun/pet-shop-server/controllers"
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
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
		c.JSON(401, models.Result{Code: 10001, Message: "token is invalid"})
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

	r.POST("/sendSignupEmailOtp", controllers.SendSignupEmailOtp)

	r.POST("/sendResetEmailOtp", controllers.SendResetEmailOtp)

	r.POST("/findbackPassword", controllers.FindbackPassword)

	r.POST("/resetPasswordByOtp", controllers.ResetPasswordByOtp)

	r.Use(checkTokenMiddleware)

	r.POST("/signOut", controllers.SignOut)

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

	r.PUT("", controllers.UpdatePet)

	r.DELETE(":PetId", controllers.DeletePet)

	r.Use(checkAdminMiddleware)

	r.POST("history", controllers.GetPetSnapShoots)
}

func getWashRecordApi(router *gin.Engine) {
	r := router.Group("/washRecord")

	r.Use(checkTokenMiddleware)
	r.Use(checkIsEmployeeMiddleware)

	r.GET(":PetId", controllers.GetPetWashRecord)

	r.POST("", controllers.CreatePetWashRecord)

	r.PUT("", controllers.UpdatePetWashRecord)

	r.DELETE(":PetId", controllers.DeletePetWashRecord)

	r.Use(checkAdminMiddleware)

	r.POST("history", controllers.GetPetWashRecordSnapShoots)
}

func getDictApi(router *gin.Engine) {
	r := router.Group("/dict")

	r.Use(checkTokenMiddleware)
	r.Use(checkIsEmployeeMiddleware)

	r.GET(":key", controllers.GetDict)
}

func getCommonApi(router *gin.Engine) {
	r := router.Group("/common")

	r.Use(checkTokenMiddleware)

	r.POST("uploadPic", controllers.CommonUploadPic)
}

func setCros(router *gin.Engine) {
	router.Use(CORSMiddleware())
}

func getLocalOss(router *gin.Engine) {
	router.Static("/uploads", "./uploads")
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	setCros(router)
	getAuthApi(router)
	getUserApi(router)
	getDictApi(router)
	getPetApi(router)
	getWashRecordApi(router)
	getCommonApi(router)
	getLocalOss(router)
	return router
}
