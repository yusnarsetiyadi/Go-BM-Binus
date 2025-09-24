package database

import (
	"bm_binus/internal/config"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var dbConnections map[string]*gorm.DB

func Init() {

	dbConfigurations := map[string]Db{
		"MYSQL": &dbMySQL{
			db: db{
				Host: config.Get().DB.DbHost,
				User: config.Get().DB.DbUser,
				Pass: config.Get().DB.DbPass,
				Port: config.Get().DB.DbPort,
				Name: config.Get().DB.DbName,
				Ssl:  config.Get().DB.DbSsl,
				Tz:   config.Get().DB.DbTz,
			},
		},
	}

	dbConnections = make(map[string]*gorm.DB)
	for k, v := range dbConfigurations {
		db, err := v.Init()
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to database %s", k))
		}
		dbConnections[k] = db
		logrus.Info(fmt.Sprintf("Successfully connected to %s", k))
	}
}

func Connection(name string) (*gorm.DB, error) {
	if dbConnections[strings.ToUpper(name)] == nil {
		return nil, errors.New("Connection is undefined")
	}
	return dbConnections[strings.ToUpper(name)], nil
}
