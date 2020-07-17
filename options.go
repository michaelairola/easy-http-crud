package easyHttpCrud

import (
	"github.com/jinzhu/gorm"
	"net/url"
	"strconv"
)

type Preload struct {
	Field    string
	Function func(*gorm.DB) *gorm.DB
}

func establishPreloads(tx *gorm.DB, record interface{}, opts map[string]interface{}) *gorm.DB {
	db := tx.Set("gorm:auto_preload", true)
	if isHistory(record) {
		return db
	}
	preloads := opts["Preloads"]
	if preloads != nil {
		loads1, ok := preloads.([]Preload)
		if ok {
			for _, load := range loads1 {
				db = db.Preload(load.Field, load.Function)
			}
		} else {
			loads2, ok := preloads.(map[string]interface{})
			if ok {
				for load, condition := range loads2 {
					if load == "AutoPreload" && condition == false {
						db = db.Set("gorm:auto_preload", false)
						continue
					}
					db = db.Preload(load, condition)
				}
			}
		}
	}
	return db
}

func establishTxScope(tx *gorm.DB, record interface{}, opts Options) *gorm.DB {
	db := tx
	debug, ok := opts["Debug"].(bool)
	if ok && debug {
		db = db.Debug()
	}
	db = establishPreloads(db, record, opts)
	omits, ok := opts["Omit"].([]string)
	if ok {
		for _, o := range omits {
			db = db.Omit(o)
		}
	}
	return db
}

func mergeOptsWithQueries(opts Options) Options {
	newOpts := CopyOpts(opts)
	queries, ok := newOpts["queries"].(url.Values)
	if ok {
		v := queries["Debug"]
		if len(v) != 0 {
			debug, _ := strconv.ParseBool(v[0])
			newOpts["Debug"] = debug
			delete(queries, "Debug")
		}
		v = queries["Omit"]
		if len(v) != 0 {
			newOpts["Omit"] = v
			delete(queries, "Omit")
		}
		newOpts["queries"] = queries
	}
	return newOpts
}
