package controllers

import (
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

func GetPetWashRecord(c *gin.Context) {
	cid := c.Param("PetId")
	var petDetail models.PetWashRecord
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

type CreatePetWashRecordRequest struct {
	models.PetWashRecord
}

func CreatePetWashRecord(c *gin.Context) {
	var request CreatePetWashRecordRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	var petWashRecord models.PetWashRecord

	petWashRecord.PetId = request.PetId
	petWashRecord.PetWashRecordId = "PetWashRecord-" + uuid.New().String()
	petWashRecord.Aggressive = request.Aggressive
	petWashRecord.IsNeedRestriction = request.IsNeedRestriction
	petWashRecord.ShapooProportion = request.ShapooProportion
	petWashRecord.SpecialRequirement = request.SpecialRequirement
	petWashRecord.BeautyRequirement = request.BeautyRequirement

	models.CommonCreate[models.PetWashRecord](&petWashRecord, c)

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&petWashRecord).Error; err != nil {
			return err
		}
		cid, _ := c.Get("cid")
		snapshoot := models.PetWashRecordSnapShoot{
			PetWashRecord: petWashRecord,
			SnapId:        "Snap-" + uuid.New().String(),
			Type:          "0",
			Editor:        cid.(string),
			PetId:         petWashRecord.PetId,
			EditTime:      time.Now(),
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
	c.JSON(200, models.Result{0, "success", petWashRecord})
}

func UpdatePetWashRecord(c *gin.Context) {
	var request models.PetWashRecord
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	petId := request.PetId
	var oldPetWashRecord models.PetWashRecord
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

	update := models.PetWashRecord{
		PetId:              request.PetId,
		PetWashRecordId:    request.PetWashRecordId,
		Aggressive:         request.Aggressive,
		IsNeedRestriction:  request.IsNeedRestriction,
		ShapooProportion:   request.ShapooProportion,
		SpecialRequirement: request.SpecialRequirement,
		BeautyRequirement:  request.BeautyRequirement,
	}

	updateCh := make(chan string)
	go models.CommonUpdate[models.PetWashRecord](&update, c, updateCh)
	<-updateCh

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pet_id = ?", petId).Updates(&update).Error; err != nil {
			return err
		}

		petWashRecord := models.PetWashRecord{}
		if err := tx.Where("pet_id = ?", petId).First(&petWashRecord).Error; err != nil {
			return err
		}

		cid, _ := c.Get("cid")
		snapshoot := models.PetWashRecordSnapShoot{
			PetWashRecord: petWashRecord,
			SnapId:        "Snap-" + uuid.New().String(),
			Type:          "1",
			Editor:        cid.(string),
			PetId:         petWashRecord.PetId,
			EditTime:      time.Now(),
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

func DeletePetWashRecord(c *gin.Context) {
	var request models.PetWashRecord
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 显示自定义的错误信息
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	petId := request.PetId
	var oldPetWashRecord models.PetWashRecord
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

type GetPetWashRecordSnapShootsReq struct {
	PageSize int    `json:"PageSize"`
	PageNo   int    `json:"PageNo"`
	PetId    string `json:"PetId"`
}

func GetPetWashRecordSnapShoots(c *gin.Context) {
	req := GetPetWashRecordSnapShootsReq{
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
