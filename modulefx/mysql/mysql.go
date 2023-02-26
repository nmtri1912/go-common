package mysql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dlmiddlecote/sqlstats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewDB(lifecycle fx.Lifecycle) *gorm.DB {
	username := viper.GetString("mysql.username")
	password := viper.GetString("mysql.password")
	url := viper.GetString("mysql.url")
	mysqlSchema := viper.GetString("mysql.schema")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s", username, password, url, mysqlSchema)

	log.Println("Connecting to database")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		log.Fatal("Cannot init mysql connection ", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("Can not get database connection")
	}

	// Maximum Idle Connections
	sqlDb.SetMaxIdleConns(12)
	// Maximum Open Connections
	sqlDb.SetMaxOpenConns(24)
	// Idle Connection Timeout
	sqlDb.SetConnMaxIdleTime(600000 * time.Millisecond)
	// Connection Lifetime
	sqlDb.SetConnMaxLifetime(1800000 * time.Millisecond)

	lifecycle.Append(fx.Hook{OnStop: func(ctx context.Context) error {
		log.Println("Closing DB")
		return sqlDb.Close()
	}})

	// Register stats with Prometheus
	collector := sqlstats.NewStatsCollector(mysqlSchema, sqlDb)
	prometheus.MustRegister(collector)

	log.Println("Connect to database successfully")
	return db
}
