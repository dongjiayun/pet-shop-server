package models

type Customer struct {
	Model
	CustomerId string `json:"customerId" gorm:"primaryKey"`
	Nickname   string `json:"nickname"`
}

func (Customer) TableName() string {
	return "customer"
}
