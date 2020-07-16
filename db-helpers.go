package easyHttpCrud

import (
	"errors"
	"net/url"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
)

// tries to connect to db indefinitely every second on the second
func InitConnectDB(connection string) (*gorm.DB, error) {
	var db *gorm.DB
	startTime := time.Now()
	GiveUpTime := float64(60 * 2)
	err := errors.New("Connecting...")
	for err != nil && time.Now().Sub(startTime).Seconds() < GiveUpTime {
		db, err = gorm.Open("mysql", connection)
		if err != nil {
			time.Sleep(time.Second)
		}
	}
	return db, err
}

func checkForUnscoped(tx *gorm.DB, queries url.Values) *gorm.DB {
	// if status is void, that means we need to search through all record items that have been deleted
	for key, matches := range queries {
		if key == "Status" || key == "status" {
			for _, match := range matches {
				if match == "Void" {
					return tx.Unscoped()
				}
			}
		}
	}
	return tx
}

func CheckForUnique(tx *gorm.DB, fields []string, record interface{}) []error {
	errs := make([]error, 0)
	table_name := TableNameOf(record)
	Id, _ := GetId(record)
	for _, Key := range fields {
		var cnt int
		col_name, t, hasCol := recordHasColumn(Key, record)
		if hasCol {
			charType := reflect.TypeOf("")
			if t != charType {
				errs = append(errs, errors.New("ERROR unique value searching for is not type string"))
			} else {
				Val := getField(record, Key).Interface().(string)
				tx.Table(table_name).Where("id != ? AND "+col_name+" = ?", Id, Val).Count(&cnt)
				if cnt != 0 {
					errs = append(errs, DuplicateErr{Key, Val})
				}
			}
		}
	}
	return errs
}

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
