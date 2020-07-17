package easyHttpCrud

func Init(model interface{}) error {
	var err error
	EstablishEnvironment()
	db, err := OpenDb()
	defer db.Close()
	if err != nil {
		return err
	}
	if !db.HasTable(model) {
		err = db.CreateTable(model).Error
	} else {
		err = db.AutoMigrate(model).Error
	}
	return err
}
