package controllers

import (
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GetPetsReq struct {
	PageSize int    `json:"PageSize"`
	PageNo   int    `json:"PageNo"`
	KeyWord  string `json:"keyword"`
}

func GetPets(c *gin.Context) {
	req := GetPetsReq{
		PageSize: 20,
		PageNo:   1,
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	pageNo := req.PageNo
	pageSize := req.PageSize
	var pets models.Pets

	var db *gorm.DB
	if req.KeyWord != "" {
		db = models.DB.Limit(pageSize).Offset((pageNo-1)*pageSize).Order("id desc").
			Where("deleted_at IS NULL").
			Where("nickname like ?", "%"+req.KeyWord+"%").
			Or("customerId like ?", "%"+req.KeyWord+"%").
			Find(&pets)
	} else {
		db = models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Order("id desc").
			Where("deleted_at IS NULL").
			Find(&pets)
	}

	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	var totalCount int64
	models.DB.Model(&pets).Count(&totalCount)
	list := models.GetListData[models.Pet](pets, pageNo, pageSize, totalCount)
	c.JSON(200, models.Result{0, "success", list})
}

func GetPet(c *gin.Context) {
	cid := c.Param("PetId")
	var petDetail models.Pet
	db := models.DB.Table("pet").
		Select("*").
		Where("PetId = ?", cid).
		Where("deleted_at IS NULL").
		First(&petDetail)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", petDetail})
}

type CreatePetRequest struct {
	models.Pet
	Nickname   string `json:"nickname" binding:"nickname" msg:"昵称不能为空"`
	TailNumber string `json:"tailNumber" binding:"tailNumber" msg:"尾号不能为空"`
}

func CreatePet(c *gin.Context) {
	var request CreatePetRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	var pet models.Pet

	pet.PetId = "Pet-" + uuid.New().String()
	pet.NickName = request.NickName
	pet.Birthday = request.Birthday
	pet.Avatar = request.Avatar
	pet.Breed = request.Breed
	pet.Type = request.Type
	pet.Gender = request.Gender
	pet.IsSterilized = request.IsSterilized
	pet.DiagnosisHistory = request.DiagnosisHistory
	pet.Forbiden = request.Forbiden
	pet.Aggressive = request.Aggressive

	customer := models.Customer{
		Nickname:   request.Nickname,
		CustomerId: "Customer-" + request.Nickname + "-" + request.TailNumber + "-" + "-" + uuid.New().String(),
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		pet.CustomerId = customer.CustomerId
		if err := tx.Create(&pet).Error; err != nil {
			return err
		}
		if err := tx.Create(&customer).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", pet})
}

func UpdatePet(c *gin.Context) {
	var pet models.Pet
	err := c.ShouldBindJSON(&pet)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	db := models.DB.Model(&pet).Updates(&pet)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", pet})
}

func DeletePet(c *gin.Context) {
	var pet models.Pet
	err := c.ShouldBindJSON(&pet)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	db := models.DB.Delete(&pet)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", pet})
}
