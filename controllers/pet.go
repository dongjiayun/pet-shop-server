package controllers

import (
	"time"

	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GetPetsReq struct {
	PageSize int    `json:"PageSize"`
	PageNo   int    `json:"PageNo"`
	KeyWord  string `json:"keyword"`
	Cid      string `json:"cid"`
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
			Where("nick_name like ?", "%"+req.KeyWord+"%").
			Or("customer_id like ?", "%"+req.KeyWord+"%")
	} else {
		db = models.DB.Limit(pageSize).Offset((pageNo - 1) * pageSize).Order("id desc").
			Where("deleted_at IS NULL")
	}

	if req.Cid != "" {
		db.Where("create_by = ?", req.Cid).
			Or("update_by = ?", req.Cid)
	}

	db.Find(&pets)
	if len(pets) == 0 {
		list := models.GetListData[models.Pet](models.Pets{}, pageNo, pageSize, 0)
		c.JSON(200, models.Result{0, "success", list})
		return
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
	db := models.DB.Table("pets").
		Select("*").
		Where("pet_id = ?", cid).
		Where("deleted_at IS NULL").
		First(&petDetail)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", petDetail})
}

type CreatePetRequest struct {
	models.Pet
	CustomerNickname string `json:"customerNickname"`
	TailNumber       string `json:"tailNumber"`
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
	pet.Weight = request.Weight
	pet.Avatar = request.Avatar
	pet.Breed = request.Breed
	pet.Type = request.Type
	pet.Gender = request.Gender
	pet.IsSterilized = request.IsSterilized
	pet.DiagnosisHistory = request.DiagnosisHistory
	pet.Forbiden = request.Forbiden
	pet.Aggressive = request.Aggressive
	pet.Remark = request.Remark

	var customer models.Customer
	var hasCustomer bool
	models.DB.Table("customer").Where("customer_id like ?", "%"+request.CustomerNickname+"-"+request.TailNumber+"%").First(&customer)

	if customer.CustomerId != "" {
		hasCustomer = true
		pet.CustomerId = customer.CustomerId
	} else {
		hasCustomer = false
		customer = models.Customer{
			Nickname:   request.CustomerNickname,
			CustomerId: "Customer-" + request.CustomerNickname + "-" + request.TailNumber + "-" + uuid.New().String(),
		}
	}

	if !hasCustomer {
		models.CommonCreate[models.Customer](&customer, c)
	}

	models.CommonCreate[models.Pet](&pet, c)

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		pet.CustomerId = customer.CustomerId
		if err := tx.Create(&pet).Error; err != nil {
			return err
		}
		if !hasCustomer {
			if err := tx.Create(&customer).Error; err != nil {
				return err
			}
		}

		cid, _ := c.Get("cid")
		snapshoot := models.PetSnapShoot{
			Pet:      pet,
			SnapId:   "Snap-" + uuid.New().String(),
			Type:     "0",
			Editor:   cid.(string),
			PetId:    pet.PetId,
			EditTime: time.Now(),
		}

		if err := tx.Create(&snapshoot).Error; err != nil {
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
	var request models.Pet
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	petId := request.PetId
	var oldPet models.Pet
	db := models.DB.Where("pet_id = ?", petId).First(&oldPet)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	update := models.Pet{
		NickName:         request.NickName,
		Birthday:         request.Birthday,
		Avatar:           request.Avatar,
		Breed:            request.Breed,
		Type:             request.Type,
		Gender:           request.Gender,
		IsSterilized:     request.IsSterilized,
		DiagnosisHistory: request.DiagnosisHistory,
		Forbiden:         request.Forbiden,
		Aggressive:       request.Aggressive,
		Remark:           request.Remark,
	}

	updateCh := make(chan string)
	go models.CommonUpdate[models.Pet](&update, c, updateCh)
	<-updateCh

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pet_id = ?", petId).Updates(&update).Error; err != nil {
			return err
		}

		pet := models.Pet{}
		if err := tx.Where("pet_id = ?", petId).First(&pet).Error; err != nil {
			return err
		}

		cid, _ := c.Get("cid")
		snapshoot := models.PetSnapShoot{
			Pet:      pet,
			SnapId:   "Snap-" + uuid.New().String(),
			Type:     "1",
			Editor:   cid.(string),
			PetId:    petId,
			EditTime: time.Now(),
		}

		if err := tx.Create(&snapshoot).Error; err != nil {
			return err
		}

		return nil
	})

	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", petId})
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

type GetPetSnapShootsReq struct {
	PageSize int    `json:"PageSize"`
	PageNo   int    `json:"PageNo"`
	PetId    string `json:"PetId"`
}

func GetPetSnapShoots(c *gin.Context) {
	req := GetPetSnapShootsReq{
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
	var pets models.PetSnapShots

	var db *gorm.DB
	if req.PetId != "" {
		db = models.DB.Limit(pageSize).Offset((pageNo-1)*pageSize).Order("id desc").
			Where("pet_id = ?", req.PetId).Find(&pets)
	} else {
		c.JSON(200, models.Result{Code: 10002, Message: "请传入PetId"})
		return
	}

	if len(pets) == 0 {
		list := models.GetListData[models.Pet](models.Pets{}, pageNo, pageSize, 0)
		c.JSON(200, models.Result{0, "success", list})
		return
	}

	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	var totalCount int64
	models.DB.Model(&pets).Count(&totalCount)
	list := models.GetListData[models.PetSnapShoot](pets, pageNo, pageSize, totalCount)
	c.JSON(200, models.Result{0, "success", list})
}
