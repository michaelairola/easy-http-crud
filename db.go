package easyHttpCrud

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/* -------------------------------------------------
				DATABASE FUNCTIONS
   -------------------------------------------------
	this is a collection of assorted database functions.
	primary used in CRUD.go to create the dymanic crud routes that every
	microservice should have, they are essential to a dynamic and robust
	microservice architecture

	all these point to tx functions in tx.go
*/

type Result map[string]interface{}
type Options map[string]interface{}

type DbFunc func(interface{}, Options) (Result, error)

func TxToDbConverter(fn TxFunc) DbFunc {
	return func(record interface{}, opts Options) (Result, error) {
		result := make(Result)
		var db *gorm.DB
		var err error
		db, err = OpenDb()
		if err != nil {
			fmt.Println("DB CONNECTION: ", err)
			return result, errors.New("There was a problem connecting to the database. Please try again in 10-15 minutes.")
		}
		defer db.Close()
		tx := db.Begin()
		result, err = fn(tx, record, opts)
		if err != nil {
			tx.Rollback()
			return result, err
		}
		tx.Commit()
		return result, nil
	}
}

func GetListByQuery(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(GetListByQueryTx)(record, opts)
}
func GetHistListByQuery(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(GetHistListByQueryTx)(record, opts)
}
func GetById(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(GetByIdTx)(record, opts)
}
func GetHistory(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(GetHistoryTx)(record, opts)
}
func Update(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(UpdateTx)(record, opts)
}
func Create(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(CreateTx)(record, opts)
}
func DeleteById(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(DeleteByIdTx)(record, opts)
}
func MergeRecords(record interface{}, opts Options) (Result, error) {
	return TxToDbConverter(MergeRecordsTx)(record, opts)
}
