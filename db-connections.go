package easyHttpCrud

import (
	// "fmt"
	"github.com/jinzhu/gorm"
	"strconv"
)

type DbConn struct {
	Host       string
	Port       int
	SqlDialect string
	User       string
	Password   string
	DbName     string
}

func (dbConn DbConn) toStr() string {
	portStr := strconv.Itoa(dbConn.Port)
	return dbConn.User + ":" + dbConn.Password + "@tcp(" + dbConn.Host + ":" + portStr + ")/" + dbConn.DbName + "?parseTime=True&charset=utf8&loc=Local"
}

func OpenDb() (*gorm.DB, error) {
	return gorm.Open(dbConn.SqlDialect, dbConn.toStr())
}
