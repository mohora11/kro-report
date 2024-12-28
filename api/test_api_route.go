package api

import (
	"github.com/gin-gonic/gin"
)

// 패키지 로드시 초기화
func init() {

}

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.Engine) {

	// 기본 path : /cafe
	testAPI := r.Group("/krononlabs")
	{
		APIApplyRoutes(testAPI)
	}

}
