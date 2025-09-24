package database

import (
	"bm_binus/internal/config"
	"fmt"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Db interface {
	Init() (*gorm.DB, error)
}

type db struct {
	Host string
	User string
	Pass string
	Port string
	Name string
	Ssl  string
	Tz   string
}

type dbMySQL struct {
	db
}

func (c *dbMySQL) Init() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s&tls=%s", c.User, c.Pass, c.Host, c.Port, c.Name, c.Tz, c.Ssl)

	var level logger.LogLevel = 4
	if gormLevel, _ := strconv.Atoi(config.Get().Logging.GormLevel); gormLevel != 0 {
		switch gormLevel {
		case 1:
			level = 1
		case 2:
			level = 2
		case 3:
			level = 3
		case 4:
			level = 4
		}
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
