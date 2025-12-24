package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局数据库对象
var DB *gorm.DB

// 初始化数据库的连接
func InitDB() {
	//数据库连接字符串
	//dsn := "root:123456@tcp(127.0.0.1:3306)/mud_game?charset=utf8mb4&parseTime=True&loc=Local"
	//dsn := "root:123456@tcp(127.0.0.1:3307)/mud_game?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "root:123456@tcp(mysql:3306)/mud_game?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	//尝试连接 10 次，每次间隔2秒
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("数据库连接成功！")
			return //连接成功，直接返回
		}

		fmt.Printf("数据库连接失败 (第 %d 次重试): %v\n", i+1, err)
		fmt.Println("正在等待 MySQL 启动...")
		time.Sleep(2 * time.Second) //等待 2 秒
	}

	// 如果 10 次都失败了，再报错退出
	panic("彻底放弃：数据库连接失败，请检查 MySQL 是否启动！")
}
