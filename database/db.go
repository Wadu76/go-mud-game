package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//全局数据库对象
var DB *gorm.DB

//初始化数据库的连接
func InitDB() {
	//数据库连接字符串
	dsn := "root:123456@tcp(127.0.0.1:3306)/mud_game?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败！"+ err.Error()) //直接崩溃
	}
	fmt.Println("数据库连接成功！")
}