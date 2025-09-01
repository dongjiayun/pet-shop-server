package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/goccy/go-json"
)

type PetWashRecord struct {
	Model
	PetWashRecordId    string      `json:"petWashRecordId"`
	PetId              string      `json:"petId"`
	Aggressive         string      `json:"aggressive"`        // 0 无 1 有
	IsNeedRestriction  string      `json:"isNeedRestriction"` // 0 无 1 有
	ShapooProportion   string      `json:"shapooProportion"`
	SpecialRequirement string      `json:"specialRequirement"`
	BeautyRequirement  string      `json:"beautyRequirement"`
	Others             string      `json:"others"`
	Attachments        Attachments `json:"attachments"`
}

func (petWashRecord *PetWashRecord) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petWashRecord = PetWashRecord{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petWashRecord)
}

func (petWashRecord PetWashRecord) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petWashRecord)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type PetWashRecordSnapShoot struct {
	Id            uint          `json:"-" gorm:"primary_key"`
	PetWashRecord PetWashRecord `json:"petWashRecord"`
	SnapId        string        `json:"snapId"`
	Type          string        `json:"type"` // 0 创建 1 修改
	Editor        string        `json:"editor"`
	PetId         string        `json:"petId"`
	EditTime      time.Time     `json:"editTime"`
	EditorName    string        `json:"editName"`
	EditorEmail   string        `json:"editEmail"`
}

func (petWashRecordSnapShoot *PetWashRecordSnapShoot) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petWashRecordSnapShoot = PetWashRecordSnapShoot{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petWashRecordSnapShoot)
}

func (petWashRecordSnapShoot PetWashRecordSnapShoot) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petWashRecordSnapShoot)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type PetWashRecordSnapShots []PetWashRecordSnapShoot

func (petWashRecordSnapShoots *PetWashRecordSnapShots) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*petWashRecordSnapShoots = PetWashRecordSnapShots{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, petWashRecordSnapShoots)
}

func (petWashRecordSnapShoots PetWashRecordSnapShots) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(petWashRecordSnapShoots)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}
