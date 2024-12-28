package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"krononlabs/api"
)

var DB *gorm.DB

func main() {

	// DB 연결
	dsn := "root:z1s2c3f4##@tcp(host.docker.internal:3306)/kronon_db?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Gin 기본 웹 서버
	ginEngine := gin.Default()

	// API 라우팅 설정
	api.ApplyRoutes(ginEngine)

	// 서버 실행(로컬)
	ginEngine.Run(":8080")

}
