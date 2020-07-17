package easyHttpCrud

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type IsValid interface {
	IsValid() bool
}
type SearchResponse struct {
	Models    interface{}
	PageCount int
	Warnings  []string
	Error     string
}

func searchFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := NewRecord(model)
		queries := r.URL.Query()
		opts["queries"] = queries
		var res Result
		var err error
		res, err = GetListByQuery(record, opts)
		var status = 200
		if err != nil {
			status = 400
			res["Error"] = err.Error()
		}
		SendObject(w, status, res)
	})
}

func getByIdFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		record := NewRecord(model)
		opts["queries"] = r.URL.Query()
		Id, _ := strconv.Atoi(mux.Vars(r)["Id"])
		opts["Id"] = Id
		_, err := GetById(record, opts)
		if err != nil {
			SendM(w, 400, "record with Id "+strconv.Itoa(Id)+" not found in table "+TableNameOf(model))
			return
		} else if err != nil {
			SendM(w, 400, "Error getting record:"+err.Error())
		}
		SendObject(w, 200, record)
	}
}
func createFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := make(Result)
		record := NewRecord(model)
		_, err := ParseBody(r, record)
		if err != nil {
			fmt.Println("Create Parsing Error: ", err.Error())
			res["Error"] = "Something went wrong. Please try again in a few minutes."
			SendObject(w, 400, res)
			return
		}
		validate, ok := record.(IsValid)
		if ok && !validate.IsValid() {
			res["Error"] = "Please make sure you have set all required values before try again."
			SendObject(w, 400, res)
			return
		}
		opts["queries"] = r.URL.Query()
		if res, err = Create(record, opts); err != nil {
			res["Error"] = err.Error()
			SendObject(w, 400, res)
			return
		}
		SendObject(w, 201, res)
	}
}

func updateFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := make(Result)
		record := NewRecord(model)
		_, err := ParseBody(r, record)
		if err != nil {
			fmt.Println("Update Parsing Error:", err.Error())
			res["Error"] = "Something went wrong. Please try again in a few minutes."
			SendObject(w, 400, res)
			return
		}
		validate, ok := record.(IsValid)
		if ok && !validate.IsValid() {
			res["Error"] = "Please make sure you have set all required values before trying again."
			SendObject(w, 400, res)
			return
		}
		Id, err := strconv.Atoi(mux.Vars(r)["Id"])
		if err != nil {
			fmt.Println("Update getting Id:", err.Error())
			res["Error"] = "Something went wrong. Please refresh page and try again."
			SendObject(w, 400, res)
			return
		}
		opts["Id"] = Id
		opts["queries"] = r.URL.Query()
		res, err = Update(record, opts)
		if err != nil {
			res["Error"] = err.Error()
			SendObject(w, 400, res)
			return
		}
		SendObject(w, 200, res)
	}
}

func deleteFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		record := NewRecord(model)
		Id, err := strconv.Atoi(mux.Vars(r)["Id"])
		if err != nil {
			SendM(w, 400, "error converting Id: "+err.Error())
			return
		}
		opts["Id"] = Id
		opts["queries"] = r.URL.Query()
		_, err = DeleteById(record, opts)
		if err != nil {
			SendM(w, 400, "Error delete record:"+err.Error())
			return
		}
		SendObject(w, 200, record)
	}
}
func mergeFactory(model interface{}, opts map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		record := NewRecord(model)
		Id1, err := strconv.Atoi(mux.Vars(r)["Id1"])
		if err != nil {
			SendM(w, 400, err.Error())
			return
		}
		Id2, err := strconv.Atoi(mux.Vars(r)["Id2"])
		if err != nil {
			SendM(w, 400, err.Error())
			return
		}
		opts["Id1"] = Id1
		opts["Id2"] = Id2
		opts["queries"] = r.URL.Query()
		_, err = MergeRecords(record, opts)
		if err != nil {
			SendM(w, 400, err.Error())
			return
		}
		SendObject(w, 200, record)
	}
}
