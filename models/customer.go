package models

type Customer struct {
	Model
	CustomerId string `json:"customerId" gorm:"primaryKey"`
	Phone      string `json:"phone" gorm:"unique"`
	Nickname   string `json:"nickname"`
}

func (Customer) TableName() string {
	return "customer"
}
