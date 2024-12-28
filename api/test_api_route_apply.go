package api

import (
	"github.com/gin-gonic/gin"
)

// APIApplyRoutes : api applies router to the gin Engine
func APIApplyRoutes(r *gin.RouterGroup) {

	r.POST("/klineinq", KlineInqReq) // Kline 조회 요청

}
