package models

import (
	"database/sql/driver"
	"errors"
	"github.com/goccy/go-json"
	"time"
)

type PetEntrustment struct {
	Model
	PetEntrustmentId   string   `json:"petEntrustmentId"`
	PetId              string   `json:"petId"`
	Habit              string   `json:"habit"`           // 0 无 1 有
	FoodRequirement    string   `json:"foodRequirement"` // 0 无 1 有
	StrollRequirement  string   `json:"strollRequirement"`
	NursingRequirement string   `json:"nursingRequirement"`
	RoomRequirement    string   `json:"roomRequirement"`
	Cautions           string   `json:"cautions"`
	SpecialRequirement string   `json:"specialRequirement"`
	Others             string   `json:"others"`
	Attachments        []string `json:"attachments"`
}

func (petEntrustment *PetEntrustment) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petEntrustment = PetEntrustment{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petEntrustment)
}

func (petEntrustment PetEntrustment) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petEntrustment)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type PetEntrustmentSnapShoot struct {
	Id             uint           `json:"-" gorm:"primary_key"`
	PetEntrustment PetEntrustment `json:"petEntrustment"`
	SnapId         string         `json:"snapId"`
	Type           string         `json:"type"` // 0 创建 1 修改
	Editor         string         `json:"editor"`
	PetId          string         `json:"petId"`
	EditTime       time.Time      `json:"editTime"`
}

func (petEntrustmentSnapShoot *PetEntrustmentSnapShoot) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petEntrustmentSnapShoot = PetEntrustmentSnapShoot{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petEntrustmentSnapShoot)
}

func (petEntrustmentSnapShoot PetEntrustmentSnapShoot) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petEntrustmentSnapShoot)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type PetEntrustmentSnapShots []PetEntrustmentSnapShoot

func (petEntrustmentSnapShoots *PetEntrustmentSnapShots) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petEntrustmentSnapShoots = PetEntrustmentSnapShots{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petEntrustmentSnapShoots)
}

func (petEntrustmentSnapShoots PetEntrustmentSnapShots) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petEntrustmentSnapShoots)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}
