package infra

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// err = db.AutoMigrate(&model.User{})
	// if err != nil {
	// 	return nil, err
	// }

	return db, nil
}
