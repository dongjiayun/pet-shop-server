package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/goccy/go-json"
)

type Breed struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func (breed *Breed) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*breed = Breed{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, breed)
}

func (breed Breed) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(breed)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Pet struct {
	Model
	PetId            string `json:"petId"`
	NickName         string `json:"nickName"`
	Birthday         string `json:"birthday"`
	Weight           string `json:"weight"`
	Avatar           string `json:"avatar"`
	Breed            Breed  `json:"breed" gorm:"json""`
	Type             string `json:"type"`         //0 猫 //1 狗 //2 其他
	Gender           string `json:"gender"`       // 0 雄性 1 雌性
	IsSterilized     string `json:"isSterilized"` // 0 未绝育 1 已绝育
	DiagnosisHistory string `json:"diagnosisHistory"`
	Forbiden         string `json:"forbiden"`
	Aggressive       string `json:"aggressive"` // 0 无 1 有
	CustomerId       string `json:"customerId"`
	Remark           string `json:"remark"`
}

func (pet *Pet) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pet = Pet{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pet)
}

func (pet Pet) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pet)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type Pets []Pet

func (pets *Pets) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pets = Pets{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pets)
}

func (pets Pets) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pets)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

// 快照

type PetSnapShoot struct {
	Id       uint      `json:"-" gorm:"primary_key"`
	Pet      Pet       `json:"pet"`
	SnapId   string    `json:"snapId"`
	Type     string    `json:"type"` // 0 创建 1 修改
	Editor   string    `json:"editor"`
	PetId    string    `json:"petId"`
	EditTime time.Time `json:"editTime"`
}

func (pet *PetSnapShoot) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pet = PetSnapShoot{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pet)
}

func (pet PetSnapShoot) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pet)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type PetSnapShots []PetSnapShoot

func (pets *PetSnapShots) Scan(value interface{}) error {
	// 将数据库中的值解析为字符串切片
	if value == nil {
		*pets = PetSnapShots{}
		return nil
	}
	stringValue, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid value type")
	}
	return json.Unmarshal(stringValue, pets)
}

func (pets PetSnapShots) Value() (driver.Value, error) {
	// 将字符串切片转换为JSON字符串存储到数据库中
	jsonString, err := json.Marshal(pets)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}
