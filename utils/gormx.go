package utils

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var client *gorm.DB

func NewGorm(host, u, p, db string, isencrypt bool) error {
	var err error
	var userx string
	var passx string
	var user = u
	var pass = p

	if isencrypt {
		userx, err = AesDecrypt(u)
		if err != nil {
			return err
		}
		passx, err = AesDecrypt(p)
		if err != nil {
			return err
		}

		user = userx
		pass = passx
	}

	var dialector gorm.Dialector
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, db)
	dialector = mysql.Open(connectionString)
	gc := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	client, err = gorm.Open(dialector, gc)
	if err != nil {
		return err
	}

	sqlDB, err := client.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Duration(10) * time.Second)

	return nil
}

func GetDB() *gorm.DB {
	return client
}
