package easyHttpCrud

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"reflect"
)

func createConn(user, password, port, db_name string) string {
	return user + ":" + password + "@tcp(" + SqlProxy + ":" + port + ")/" + db_name + "?parseTime=True&charset=utf8&loc=Local"
}

func OpenDb() (*gorm.DB, error) {
	conn, err := Conn()
	if err != nil {
		return nil, err
	}
	return OpenDbWithString(conn)
}

func OpenDbFromReq(req *http.Request) (*gorm.DB, error) {
	reqConnStr, err := ConnFromReq(req)
	if err != nil {
		return nil, err
	}
	return OpenDbWithString(reqConnStr)
}

// for no log of callbacks
type NoLog struct{}

func (NoLog) Print(values ...interface{}) {}

type DbFn func(db *gorm.DB) error

func WrapNoLog(db *gorm.DB, dbFn func(db *gorm.DB) error) error {
	db.SetLogger(NoLog{})
	err := dbFn(db)
	db.SetLogger(gorm.Logger{log.New(os.Stdout, "\r\n", 0)})
	return err
}

func OpenDbWithString(connection_string string) (*gorm.DB, error) {
	db, err := gorm.Open("mysql", connection_string)
	if err != nil {
		fmt.Println("err connecting to db:", err)
		return db, err
	}
	WrapNoLog(db, func(db *gorm.DB) error {
		db.Callback().Create().Before("gorm:before_create").Register("defie:init_id", initializeId)
		db.Callback().Update().Before("gorm:update").Register("defie:update_ver", updateVer)
		db.Callback().Query().After("gorm:after_query").Register("defie:populate_idver", populateIdVer)
		db.Callback().Create().After("gorm:create").Register("defie:populate_idver", populateIdVer)
		db.Callback().Update().After("gorm:update").Register("defie:populate_idver", populateIdVer)
		return nil
	})
	return db, err
}

func callGormHook(hook string) func(interface{}, *gorm.DB) error {
	return func(model interface{}, tx *gorm.DB) error {
		method := reflect.ValueOf(model).MethodByName(hook)
		if method.Kind() != reflect.Invalid {
			txV := reflect.ValueOf(tx)
			v := method.Call([]reflect.Value{txV})[0].Interface()
			err, ok := v.(error)
			if v != nil && !ok {
				return errors.New("gorm hook " + hook + " does not return an error.")
			}
			return err
		}
		return nil
	}
}
func callBeforeUpdate(model interface{}, tx *gorm.DB) error {
	return callGormHook("BeforeUpdate")(model, tx)
}
func callBeforeSave(model interface{}, tx *gorm.DB) error {
	return callGormHook("BeforeSave")(model, tx)
}
func callBefore(model interface{}, tx *gorm.DB) error {
	err := callBeforeUpdate(model, tx)
	if err != nil {
		fmt.Println("before update gorm hook error")
		return err
	}
	err = callBeforeSave(model, tx)
	if err != nil {
		fmt.Println("before save gorm hook error")
		return err
	}
	return nil
}
