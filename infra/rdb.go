package infra

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"

	// Just import.
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	// RDB provides gorm db reference.
	RDB *gorm.DB
)

type dbConfig struct {
	Endpoint string
	Database string
	Username string
	Password string
}

var dbConfigs = map[environment]dbConfig{
	dev: dbConfig{
		Database: "slgmails_dev",
	},
	stg: dbConfig{
		Database: "slgmails_stg",
	},
	prod: dbConfig{
		Database: "slgmails_prod",
	},
}

func setupDatabase() {
	// Determine base config per env.
	dbc := dbConfigs[getEnvironment()]

	log.Println("Get RDS username & password from env...")
	dbc.Username = Env.MysqlUser
	dbc.Password = Env.MysqlPass

	log.Println("Get MYSQL endpoint...")

	dbc.Endpoint = Env.MysqlEndpoint

	if dbc.Endpoint == "" {
		panic("[Error] db endpoint not found on environment variable and config")
	}

	log.Println("Connect database...")
	connectToDatabase(
		fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			dbc.Username, dbc.Password, dbc.Endpoint, dbc.Database),
	)
}

func connectToDatabase(dns string) {
	db, err := gorm.Open("mysql", dns)
	if err != nil {
		panic(fmt.Sprintf("[Error] %s", err))
	}

	db.LogMode(!isProduction())

	// SetMaxIdleConns() should be greater than SetMaxOpenConns().
	db.DB().SetMaxIdleConns(30)
	db.DB().SetMaxOpenConns(30)
	// SetConnMaxLifetime() should be MaxOpenConns * 1sec.
	db.DB().SetConnMaxLifetime(time.Second * 30)
	RDB = db
}
