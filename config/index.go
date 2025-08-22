package config

import "time"

const Secret = "123456"

const DataBase = "root:@tcp(127.0.0.1:3306)/pet-shop-app?charset=utf8mb4&parseTime=True&loc=Local"

//const DataBase = "root:@tcp(1.94.65.197:3306)/pet-family?charset=utf8mb4&parseTime=True&loc=Local"

const SmtpHost = "smtp.163.com"

const SmtpPort = 465

const SmtpUser = "birkinpet@163.com"

const TokenExpireDuration = time.Hour * 24 * 30

// OBS credentials (replace with actual values)
const ObsAK = "your-obs-access-key"
const ObsSK = "your-obs-secret-key"
