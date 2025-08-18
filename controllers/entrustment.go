package controllers

import (
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

func GetPetEntrustment(c *gin.Context) {
	cid := c.Param("PetId")
	var petDetail models.PetEntrustment
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

type CreatePetEntrustmentRequest struct {
	models.PetEntrustment
}

func CreatePetEntrustment(c *gin.Context) {
	var request CreatePetEntrustmentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	var petEntrustment models.PetEntrustment

	petEntrustment.PetId = request.PetId
	petEntrustment.PetEntrustmentId = "petEntrustment-" + uuid.New().String()

	petEntrustment.Habit = request.Habit
	petEntrustment.FoodRequirement = request.FoodRequirement
	petEntrustment.StrollRequirement = request.StrollRequirement
	petEntrustment.NursingRequirement = request.NursingRequirement
	petEntrustment.RoomRequirement = request.RoomRequirement
	petEntrustment.Cautions = request.Cautions
	petEntrustment.SpecialRequirement = request.SpecialRequirement
	petEntrustment.Others = request.Others
	petEntrustment.Attachments = request.Attachments

	models.CommonCreate[models.PetEntrustment](&petEntrustment, c)

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&petEntrustment).Error; err != nil {
			return err
		}
		cid, _ := c.Get("cid")
		snapshoot := models.PetEntrustmentSnapShoot{
			PetEntrustment: petEntrustment,
			SnapId:         "Snap-" + uuid.New().String(),
			Type:           "0",
			Editor:         cid.(string),
			PetId:          petEntrustment.PetId,
			EditTime:       time.Now(),
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
	c.JSON(200, models.Result{0, "success", petEntrustment})
}

func UpdatePetEntrustment(c *gin.Context) {
	var request models.PetEntrustment
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	petId := request.PetId
	var oldPetWashRecord models.PetEntrustment
	db := models.DB.Where("pet_id = ?", petId).First(&oldPetWashRecord)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}

	update := models.PetEntrustment{
		PetId:              request.PetId,
		PetEntrustmentId:   request.PetEntrustmentId,
		Habit:              request.Habit,
		FoodRequirement:    request.FoodRequirement,
		StrollRequirement:  request.StrollRequirement,
		NursingRequirement: request.NursingRequirement,
		RoomRequirement:    request.RoomRequirement,
		Cautions:           request.Cautions,
		SpecialRequirement: request.SpecialRequirement,
		Others:             request.Others,
		Attachments:        request.Attachments,
	}

	updateCh := make(chan string)
	go models.CommonUpdate[models.PetEntrustment](&update, c, updateCh)
	<-updateCh

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pet_id = ?", petId).Updates(&update).Error; err != nil {
			return err
		}

		petEntrustment := models.PetEntrustment{}
		if err := tx.Where("pet_id = ?", petId).First(&petEntrustment).Error; err != nil {
			return err
		}

		cid, _ := c.Get("cid")
		snapshoot := models.PetEntrustmentSnapShoot{
			PetEntrustment: petEntrustment,
			SnapId:         "Snap-" + uuid.New().String(),
			Type:           "1",
			Editor:         cid.(string),
			PetId:          petEntrustment.PetId,
			EditTime:       time.Now(),
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

func DeletePetEntrustment(c *gin.Context) {
	var request models.PetEntrustment
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	petId := request.PetId
	var oldPetWashRecord models.PetEntrustment
	db := models.DB.Where("pet_id = ?", petId).First(&oldPetWashRecord)
	if db.Error != nil {
		if db.Error.Error() == "record not found" {
			c.JSON(200, models.Result{Code: 10001, Message: "未找到该条记录"})
			return
		}
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	db = models.DB.Delete(&oldPetWashRecord)
	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	c.JSON(200, models.Result{0, "success", petId})
}

type GetPetEntrustmentSnapShootsReq struct {
	PageSize int    `json:"PageSize"`
	PageNo   int    `json:"PageNo"`
	PetId    string `json:"PetId"`
}

func GetPetEntrustmentSnapShoots(c *gin.Context) {
	req := GetPetEntrustmentSnapShootsReq{
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
	var snaps models.PetWashRecordSnapShots

	var db *gorm.DB
	if req.PetId != "" {
		db = models.DB.Limit(pageSize).Offset((pageNo-1)*pageSize).Order("id desc").
			Where("pet_id = ?", req.PetId).Find(&snaps)
	} else {
		c.JSON(200, models.Result{Code: 10002, Message: "请传入PetId"})
		return
	}

	if len(snaps) == 0 {
		list := models.GetListData[models.PetWashRecordSnapShoot](models.PetWashRecordSnapShots{}, pageNo, pageSize, 0)
		c.JSON(200, models.Result{0, "success", list})
		return
	}

	if db.Error != nil {
		// SQL执行失败，返回错误信息
		c.JSON(200, models.Result{Code: 10002, Message: "internal server error"})
		return
	}
	var totalCount int64
	models.DB.Model(&snaps).Count(&totalCount)
	list := models.GetListData[models.PetWashRecordSnapShoot](snaps, pageNo, pageSize, totalCount)
	c.JSON(200, models.Result{0, "success", list})
}
