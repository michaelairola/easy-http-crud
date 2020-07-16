package easyHttpCrud

import (
	"errors"
	// "fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
)

var OldRecord interface{}

type TxFunc func(*gorm.DB, interface{}, Options) (Result, error)
type hasWarnings interface {
	Warnings(*gorm.DB) []error
}
type Merger func(*gorm.DB, int, int) error

func GetListByQueryTxFactory(queryType string) func(*gorm.DB, interface{}, Options) (Result, error) {
	return func(tx *gorm.DB, model interface{}, opts Options) (Result, error) {
		result := make(map[string]interface{})
		var record interface{}
		var db *gorm.DB
		var err error
		if queryType == "history" {
			record, err = createHistoryRecord(model)
			if err != nil {
				return result, err
			}
		} else {
			record = model
		}
		records := NewRecordSlice(record)

		opts = mergeOptsWithQueries(opts)
		db = establishTxScope(tx, record, opts)
		if queryType != "history" {
			db = checkForUnscoped(db, opts["queries"].(url.Values))
		}

		queries := opts["queries"].(url.Values)
		var ParseFn func(interface{}, url.Values) (string, []interface{}, string, int, int, []error)
		if queryType == "history" {
			ParseFn = ParseHistQueryStr
		} else {
			ParseFn = ParseQueryStr
		}
		queryStr, args, order, page, pageLimit, warnings := ParseFn(model, queries)
		db = db.Table(TableNameOf(record))
		if order != "" {
			db = db.Order(order)
		}
		db = db.Where(queryStr, args...)
		var cnt int
		db.Count(&cnt)

		if pageLimit != 0 {
			offset := page * pageLimit
			db = db.Offset(offset).Limit(pageLimit)
		}
		err = db.Find(records).Error
		var pages int
		if pageLimit != 0 {
			pages = cnt/pageLimit + 1
		} else {
			pages = 1
		}
		result["Models"] = records
		result["PageCount"] = pages
		result["Count"] = cnt
		result["Warnings"] = GetMessages(warnings)
		return result, err
	}
}

var GetListByQueryTx = GetListByQueryTxFactory("")
var GetHistListByQueryTx = GetListByQueryTxFactory("history")

func GetByIdTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	result := make(Result)
	Id, ok := opts["Id"].(int)
	if !ok {
		result["Id"] = opts["Id"]
		return result, errors.New("Unable to parse Id")
	}
	db := establishTxScope(tx, record, mergeOptsWithQueries(opts))
	recordNotFound := db.Unscoped().Where("id = ?", Id).Find(record).RecordNotFound()
	if recordNotFound {
		result["Found"] = false
		result["Id"] = Id
		return result, errors.New("Record Not Found")
	}
	result["Record"] = record
	return result, nil
}
func GetHistoryTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	result := make(Result)
	Id, ok := opts["Id"].(int)
	if !ok {
		result["Id"] = opts["Id"]
		return result, errors.New("Unable to parse Id")
	}
	Ver, ok := opts["Ver"].(int)
	if !ok {
		result["Ver"] = opts["Ver"]
		return result, errors.New("Unable to parse Version")
	}
	db := establishTxScope(tx, record, mergeOptsWithQueries(opts))
	recordNotFound := db.Where("id = ? AND ver = ?", Id, Ver).Find(record).RecordNotFound()
	if recordNotFound {
		result["Found"] = false
		result["Id"] = Id
		return result, errors.New("Record Not Found")
	}
	result["Record"] = record
	return result, nil
}

func CreateTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	result := make(Result)
	db := establishTxScope(tx, record, mergeOptsWithQueries(opts))
	warn, ok := record.(hasWarnings)
	if ok {
		warnings := warn.Warnings(tx)
		result["Warnings"] = GetMessages(warnings)
	}
	OldRecord = NewRecord(record)
	err := db.Create(record).Error
	if err != nil {
		return result, err
	}
	err = saveHistory(db, record)
	if err != nil {
		return result, err
	}
	result["Record"] = record
	return result, nil
}

// TODO incorporate this baby into UpdateTx somehow... until then, history will be off whenever this func is used.
func NewUpdateTx(tx *gorm.DB, record interface{}, columnsToUpdate []string) error {
	updates := GetUpdates(record, columnsToUpdate)
	Id, err := GetId(record)
	if err != nil {
		return err
	}
	err = tx.Table(TableNameOf(record)).Where("id = ?", Id).Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	var err error
	result := make(Result)
	db := establishTxScope(tx, record, mergeOptsWithQueries(opts))
	// check if Status is Void. If so, then delete instead.
	statusField := getField(record, "Status")
	if statusField.IsValid() {
		status, _ := statusField.Interface().(string)
		if status == "Void" {
			return DeleteByIdTx(tx, record, opts)
		}
	}

	modelId, _ := GetId(record)
	Id, _ := opts["Id"].(int)
	if Id != 0 && Id != modelId {
		service := TableNameOf(record)
		err = errors.New(service + " Id is not equal to the api's Id value.")
		return result, err
	}
	warn, ok := record.(hasWarnings)
	if ok {
		warnings := warn.Warnings(tx)
		result["Warnings"] = GetMessages(warnings)
	}

	OldRecord = NewRecord(record)
	_, err = GetByIdTx(tx, OldRecord, opts)
	if err != nil {
		return result, errors.New("Record not found")
	}

	if err = db.Save(record).Error; err != nil {
		return result, err
	}
	err = saveHistory(db, record)
	if err != nil {
		return result, err
	}
	result["Record"] = record
	return result, nil
}

func DeleteByIdTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	result := make(Result)
	_, ok := opts["Id"].(int)
	if !ok {
		return result, errors.New("Error parsing Id.")
	}
	result, err := GetByIdTx(tx, record, opts)
	if err != nil {
		return result, errors.New("Record not found")
	}
	hasStatus := getField(record, "Status").IsValid()
	if hasStatus {
		err = SetField(record, "Status", "Void")
		if err != nil {
			return result, err
		}
		err = NewUpdateTx(tx, record, []string{"Status"})
		if err != nil {
			return result, err
		}
	}
	OldRecord = NewRecord(record)
	_, err = GetByIdTx(tx, OldRecord, opts)
	if err != nil {
		return result, errors.New("Record not found")
	}

	err = tx.Delete(record).Error
	if err != nil {
		return result, err
	}
	err = saveHistory(tx, record)
	if err != nil {
		return result, err
	}
	result["record"] = record
	return result, nil
}

func MergeRecordsTx(tx *gorm.DB, record interface{}, opts Options) (Result, error) {
	opts["Id"] = opts["Id1"]
	record1 := reflect.New(reflect.TypeOf(record).Elem()).Interface()
	res, err := DeleteByIdTx(tx, record1, opts)
	if err != nil {
		return res, err
	}
	merger := opts["merge"].(Merger)
	err = merger(tx, opts["Id1"].(int), opts["Id2"].(int))
	if err != nil {
		return res, err
	}
	opts["Id"] = opts["Id2"]
	res, err = GetByIdTx(tx, record, opts)
	if err != nil {
		return res, err
	}
	err = saveHistory(tx, record)
	return res, err
}

func saveHistory(db *gorm.DB, record interface{}) error {
	if hasHistory(record) {
		histRecord, err := createHistoryRecord(record)
		if err != nil {
			return err
		}
		if err = db.Save(histRecord).Error; err != nil {
			return err
		}
	}
	return nil
}
