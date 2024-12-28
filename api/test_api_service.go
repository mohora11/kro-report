package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	bsml "krononlabs/busimodel"
)

// KlineInqReq : Kline조회요청
func KlineInqReq(c *gin.Context) {

	// 요청 정보
	reqInfo := bsml.KlineRequest{}

	// json 요청정보 바인딩
	if err := c.BindJSON(&reqInfo); err != nil {
		fmt.Println(reqInfo)
		c.JSON(http.StatusBadRequest, gin.H{
			"meta": gin.H{
				"code":    http.StatusBadRequest,
				"message": "잘못된 요청입니다. 요청 형식을 확인해주세요.",
			},
		})
		return
	}

	//
	resultFlag, errMsg, resp := KlineInq(&reqInfo)

	// 정상 처리 및 오류 정보 회신
	if resultFlag {

		c.JSON(http.StatusOK, gin.H{
			"meta": gin.H{
				"code":    http.StatusOK,
				"message": errMsg,
			},
			"data": resp,
		})
	} else {

		c.JSON(http.StatusInternalServerError, gin.H{
			"meta": gin.H{
				"code":    http.StatusInternalServerError,
				"message": errMsg,
			},
			"data": resp,
		})
	}
}
