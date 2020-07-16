package easyHttpCrud

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"sync"
	"time"
)

type DbChan struct {
	Host string
	Conn string
	Db   *gorm.DB
	Err  error
}

func Init(migrations []Migration, models ...interface{}) {
	c := make(chan DbChan)
	var wg sync.WaitGroup
	for host, conn := range HostDbMap {
		go func(host string, conn string) {
			wg.Add(1)
			var dbChan DbChan
			dbChan.Host = host
			dbChan.Conn = conn
			db, err := InitConnectDB(conn)
			dbChan.Db = db
			dbChan.Err = err
			c <- dbChan
		}(host, conn)
	}
	go func() {
		for dbChan := range c {
			db := dbChan.Db
			if dbChan.Err != nil {
				fmt.Println("unable to connect to host:", dbChan.Host, dbChan.Conn, dbChan.Err.Error())
				wg.Done()
				continue
			}
			CheckForMigrationTable(db)
			tx := db.Begin()
			if err := MigrateDatabase(tx, migrations, models); err != nil {
				fmt.Println("database for host", dbChan.Host, "error:", err.Error())
				tx.Rollback()
			} else {
				fmt.Println("database for host", dbChan.Host, "successfully migrated!")
				tx.Commit()
			}

			MainModel := models[0]
			CheckForHistoryTable(db, MainModel)

			db.Close()
			wg.Done()
		}
	}()
	time.Sleep(time.Second)
	wg.Wait()
}
