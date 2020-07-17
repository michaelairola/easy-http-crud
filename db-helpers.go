package easyHttpCrud

import (
	// "errors"
	// "time"

	"github.com/jinzhu/gorm"
)

// tries to connect to db indefinitely every second on the second
// func InitConnectDB(connection string) (*gorm.DB, error) {
// 	var db *gorm.DB
// 	startTime := time.Now()
// 	GiveUpTime := float64(60 * 2)
// 	err := errors.New("Connecting...")
// 	for err != nil && time.Now().Sub(startTime).Seconds() < GiveUpTime {
// 		db, err = gorm.Open("mysql", connection)
// 		if err != nil {
// 			time.Sleep(time.Second)
// 		}
// 	}
// 	return db, err
// }

var OrderByIndex = func(db *gorm.DB) *gorm.DB {
	return db.Order("`index` ASC")
}

func checkForTable(model interface{}) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		if !tx.HasTable(model) {
			return tx.CreateTable(model).Error
		} else {
			return tx.AutoMigrate(model).Error
		}
		return nil
	}
}
