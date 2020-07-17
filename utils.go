package easyHttpCrud

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type HasTableName interface {
	TableName() string
}

func TableNameOf(model interface{}) string {
	tableNameModel, ok := model.(HasTableName)
	if ok {
		return tableNameModel.TableName()
	}
	modelV := getReflectValue(model)

	return gorm.ToTableName(modelV.Type().Name())
}

func NewRecord(record interface{}) interface{} {
	return reflect.New(reflect.TypeOf(record).Elem()).Interface()
}
func NewRecordSlice(record interface{}) interface{} {
	return reflect.New(reflect.SliceOf(reflect.TypeOf(record).Elem())).Interface()
}
func getReflectValue(model interface{}) reflect.Value {
	v := reflect.ValueOf(model)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func sliceOf(records interface{}) []interface{} {
	value := getReflectValue(records)
	switch value.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		var records []interface{}
		for i := 0; i < value.Len(); i++ {
			record := value.Index(i).Interface()
			records = append(records, record)
		}
		return records
	}
	return []interface{}{records}
}

func getField(model interface{}, field_name string) reflect.Value {
	v := getReflectValue(model)
	if v.Kind() == reflect.Struct {
		return v.FieldByName(field_name)
	} else {
		return reflect.Zero(v.Type())
	}
}
func GetFieldVal(model interface{}, fieldName string) (interface{}, error) {
	f := getField(model, fieldName)
	if !f.IsValid() {
		return 0, errors.New("No Field Named " + fieldName + " in model.")
	}
	return f.Interface(), nil
}
func GetFieldVals(model interface{}, fieldNames []string) (map[string]interface{}, error) {
	fieldVals := make(map[string]interface{})
	for _, fieldName := range fieldNames {
		v, err := GetFieldVal(model, fieldName)
		if err != nil {
			return fieldVals, err
		}
		fieldVals[fieldName] = v
	}
	return fieldVals, nil
}
func GetFieldInt(model interface{}, fieldName string) (int, error) {
	v, err := GetFieldVal(model, fieldName)
	if err != nil {
		return 0, err
	}
	n, ok := v.(int)
	if !ok {
		return 0, errors.New("Field Named " + fieldName + " Not an integer in model.")
	}
	return n, nil
}

func SetField(model interface{}, fieldName string, v interface{}) error {
	f := getField(model, fieldName)
	if !f.IsValid() {
		return errors.New("No Field Named " + fieldName + " in model.")
	}
	if !f.CanSet() {
		return errors.New("Cannot set field " + fieldName + ". model needs to be a pointer.")
	}
	fType := reflect.TypeOf(f.Interface())
	vType := reflect.TypeOf(v)
	if fType != vType {
		fmt.Println("fType:", fType, "vType:", vType)
		return errors.New("Field and value are not of same type.")
	}
	switch f.Kind() {
	case reflect.Float64:
		f.SetFloat(v.(float64))
	case reflect.Int:
		f.SetInt(int64(v.(int)))
	case reflect.String:
		f.SetString(string(v.(string)))
	}
	return nil
}
func SetFields(model interface{}, fieldVs map[string]interface{}) error {
	for fieldName, v := range fieldVs {
		err := SetField(model, fieldName, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func GetId(model interface{}) (int, error) {
	return GetFieldInt(model, "Id")
}
func hasField(model interface{}, field string) bool {
	_, err := GetFieldVal(model, field)
	if err != nil {
		return false
	}
	return true
}
func hasIdAndVer(model interface{}) bool {
	return hasField(model, "Id") && hasField(model, "Ver")
}

var hasHistory = hasIdAndVer

func SetIdField(id int, model interface{}) error {
	return SetField(model, "Id", id)
}

func DeepFields(record interface{}) []reflect.StructField {
	var structfields []reflect.StructField
	s := reflect.TypeOf(record).Elem()
	for i := 0; i < s.NumField(); i++ {
		sf := s.Field(i)
		// Anonymous == Embedded Field
		if sf.Anonymous {
			iface := reflect.New(sf.Type).Interface()
			structfields = append(structfields, DeepFields(iface)...)
		} else {
			structfields = append(structfields, sf)
		}
	}
	return structfields
}
func GetUpdates(record interface{}, ups []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, f := range DeepFields(record) {
		column_name := getColumnNameFromField(f)
		if column_name == "" {
			continue
		}
		for _, col := range ups {
			if column_name == col || f.Name == col {
				result[column_name] = getField(record, f.Name).Interface()

			}
		}
	}
	return result
}
func GetAllUpdates(record interface{}) map[string]interface{} {
	var colNames []string
	for _, strf := range DeepFields(record) {
		column_name := getColumnNameFromField(strf)
		colNames = append(colNames, column_name)
	}
	updates := GetUpdates(record, colNames)
	// shouldn't update id lolol
	delete(updates, "id")
	return updates
}

func recordHasColumn(col string, record interface{}) (string, reflect.Type, bool) {
	structfields := DeepFields(record)
	for _, f := range structfields {
		column_name := getColumnNameFromField(f)
		if col == f.Name || col == column_name {
			return "`" + column_name + "`", f.Type, true
		}
	}
	return "", nil, false
}

func getColumnNameFromField(f reflect.StructField) string {
	gormTag := f.Tag.Get("gorm")
	var columnName string
	if gormTag == "-" {
		return ""
	}
	if strings.Contains(gormTag, "column") {
		TagsArr := strings.Split(gormTag, ";")
		for _, i := range TagsArr {
			KeyVal := strings.Split(i, ":")
			if KeyVal[0] == "column" && len(KeyVal) == 2 {
				columnName = KeyVal[1]
			}
		}
	}
	if columnName == "" {
		columnName = gorm.ToColumnName(f.Name)
	}
	return columnName
}

func TermToDays(term string) int {
	termsplit := strings.Split(term, " ")
	result, err := strconv.Atoi(termsplit[0])
	if err != nil {
		return 0
	}
	return result
}

func GetMessages(errs []error) []string {
	ms := make([]string, 0)
	for _, err := range errs {
		ms = append(ms, err.Error())
	}
	return ms
}

func TimeToStr(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}
func mapInterfaceToMapInt(vs map[string]interface{}) map[string]int {
	result := make(map[string]int)
	for k, v := range vs {
		f, _ := v.(int)
		result[k] = f
	}
	return result
}
func Copy(q url.Values) url.Values {
	q_copy := make(url.Values)
	for key, v := range q {
		q_copy[key] = v
	}
	return q_copy
}
func CopyOpts(q Options) Options {
	q_copy := make(Options)
	for key, i := range q {
		switch v := i.(type) {
		case url.Values:
			q_copy[key] = Copy(v)
		default:
			q_copy[key] = v
		}
	}
	return q_copy
}

func scopeOpts(opts map[string]interface{}, key string) map[string]interface{} {
	// opts["API"] = key
	newOpts := make(map[string]interface{})
	for k, v := range opts {
		if k != key {
			newOpts[k] = v
		}
	}
	opt := opts[key]
	if opt != nil {
		opt, ok := opt.(map[string]interface{})
		if ok {
			for k, v := range opt {
				newOpts[k] = v
			}
		}
	}
	return newOpts
}

func isHistory(value interface{}) bool {
	return strings.HasSuffix(TableNameOf(value), "_history")
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func Round(x float64) int {
	return int(math.RoundToEven(x))
}
func TaxAmt(price int, rate float64) int {
	return Round(float64(price) * rate / 100)
}
