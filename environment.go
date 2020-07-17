package easyHttpCrud

import (
	// "fmt"
	"os"
	"reflect"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// PORT
	Port int

	// VERSION
	Version string
	//HOST_NAME
	HostName string

	// DEV_MODE
	DevMode bool

	dbConn = DbConn{
		Host:       "", // DB_HOST
		Port:       0,  // DB_PORT
		SqlDialect: "", // SQL_DIALECT
		User:       "", // DB_USER
		Password:   "", // DB_PASSWORD
		DbName:     "", // DB_NAME
	}
)

func EstablishEnvironment() {
	DevMode, _ = strconv.ParseBool(os.Getenv("DEV_MODE"))
	getEnvVar("HOST_NAME", &HostName, "localhost")
	getEnvVar("VERSION", &Version, "development")

	getEnvVar("DB_HOST", &dbConn.Host, "localhost")
	getEnvVar("DB_PORT", &dbConn.Port, 3306)
	getEnvVar("SQL_DIALECT", &dbConn.SqlDialect, "mysql")
	getEnvVar("DB_USER", &dbConn.User, "root")
	getEnvVar("DB_PASSWORD", &dbConn.Password, "password")
	getEnvVar("DB_NAME", &dbConn.DbName, "testing")

	getEnvVar("PORT", &Port, 3001)
}

func getEnvVar(varName string, routerVar interface{}, default_value interface{}) {
	envVarStr := os.Getenv(varName)
	varVal := getReflectValue(routerVar)
	if varVal.IsValid() && varVal.CanSet() {
		switch varVal.Kind() {
		case reflect.Int:
			var envVar int
			if envVarStr == "" {
				envVar = default_value.(int)
			} else {
				envVar, _ = strconv.Atoi(envVarStr)
			}
			varVal.SetInt(int64(envVar))
		case reflect.String:
			if envVarStr == "" {
				envVarStr = default_value.(string)
			}
			varVal.SetString(envVarStr)
		}
	}
}
