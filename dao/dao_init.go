package dao

import (
	"Main/model"
	"Main/utils"
	"log"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dataBase *gorm.DB
	once     sync.Once
)

func init() {
	once.Do(func() {
		dsn := utils.GetDataBaseUserName() + ":" +
			utils.GetDataBasePassword() + "@tcp(" +
			utils.GetDatabaseHost() + ":" +
			utils.GetDatabasePort() + ")/" +
			utils.GetDatabaseName() +
			utils.GetDataBaseExtraConfig()
		var err0 error
		dataBase, err0 = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err0 != nil {
			log.Println("数据库连接失败！")
			panic(err0)
		}
		err1 := dataBase.AutoMigrate(&model.ChatModel{})
		if err1 != nil {
			log.Println("数据库建表失败！")
			panic(err1)
		}
	})
}
