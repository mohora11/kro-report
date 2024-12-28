package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	bsml "krononlabs/busimodel"
	utility "krononlabs/utility"
)

const (
	binanceKline = "https://api.binance.com/api/v3/klines"
	bitgetKline  = "https://api.bitget.com/api/v2/spot/market/history-candles"
)

// Kline 통합 조회
func KlineInq(ReqInfo *bsml.KlineRequest) (bool, string, []bsml.KlineResp) {

	var (
		resp []bsml.KlineResp

		url string
	)

	// 거래소별 interval 값 변환
	interval := utility.GetExchangeInterval(ReqInfo.Interval, ReqInfo.ExchgNm)

	if ReqInfo.ExchgNm == "binance" {
		// 바이낸스 Kline API 요청 URL 생성
		url = fmt.Sprintf("%s?symbol=%s&interval=%s", binanceKline, ReqInfo.Symbol, interval)
	} else if ReqInfo.ExchgNm == "bitget" {
		// 비트겟 Kline API 요청 URL 생성
		url = fmt.Sprintf("%s?symbol=%s&granularity=%s", bitgetKline, ReqInfo.Symbol, interval)
	}

	// 선택적인 값이 존재하면 URL에 추가
	if ReqInfo.StartTime != "" {

		// 날짜 문자열을 time.Time 객체로 변환
		start, err := time.Parse(time.RFC3339, ReqInfo.StartTime)
		if err != nil {
			fmt.Println("Error parsing start time:", err)

			return false, "time parse error", resp
		}

		startUnix := start.Unix() * 1000

		url = fmt.Sprintf("%s&startTime=%d", url, startUnix)
	}

	if ReqInfo.EndTime != "" {

		end, err := time.Parse(time.RFC3339, ReqInfo.EndTime)
		if err != nil {
			fmt.Println("Error parsing end time:", err)

			return false, "time parse error", resp
		}

		endUnix := end.Unix() * 1000

		url = fmt.Sprintf("%s&endTime=%d", url, endUnix)
	}

	if ReqInfo.Limit > 0 {
		url = fmt.Sprintf("%s&limit=%d", url, ReqInfo.Limit)
	}

	if ReqInfo.TimeZone != "" {
		url = fmt.Sprintf("%s&timeZone=%s", url, ReqInfo.TimeZone)
	}

	// HTTP GET 요청 보내기
	httpResp, err := http.Get(url)
	if err != nil {
		return false, "Request failed", resp
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return false, "", nil
	}

	// 응답 JSON 출력
	fmt.Println("Response body:", string(body))

	// 거래소에 따라 응답을 다르게 처리
	if ReqInfo.ExchgNm == "binance" {
		// 바이낸스 응답 처리
		var rawData [][]interface{}
		if err := json.Unmarshal(body, &rawData); err != nil {
			return false, fmt.Sprintf("Error unmarshaling Binance response: %v", err), nil
		}

		// rawData를 KlineResp 구조체로 변환
		for _, kline := range rawData {
			if len(kline) < 6 {
				// 최소 6개의 필드가 있어야 함
				continue
			}

			// 공통 필드들
			timestamp := int64(kline[0].(float64))
			open := kline[1].(string)
			high := kline[2].(string)
			low := kline[3].(string)
			close := kline[4].(string)
			volume := kline[5].(string)
			closeTime := int64(kline[6].(float64))    // closeTime
			quoteVolume := kline[7].(string)          // quoteVolume 사용
			numberOfTrades := int(kline[8].(float64)) // numberOfTrades
			takerBuyBaseVolume := kline[9].(string)   // takerBuyBaseVolume
			takerBuyQuoteVolume := kline[10].(string) // takerBuyQuoteVolume

			// KlineResp 구조체에 데이터를 할당
			resp = append(resp, bsml.KlineResp{
				Timestamp:           timestamp,
				Open:                utility.ParseToFloat64(open),
				High:                utility.ParseToFloat64(high),
				Low:                 utility.ParseToFloat64(low),
				Close:               utility.ParseToFloat64(close),
				Volume:              utility.ParseToFloat64(volume),
				CloseTime:           closeTime,                                   // 바이낸스에는 CloseTime이 없으므로 0
				QuoteVolume:         utility.ParseToFloat64(quoteVolume),         // 바이낸스에서 quoteVolume은 사용하지 않으므로 0으로 처리
				NumberOfTrades:      numberOfTrades,                              // 바이낸스에는 NumberOfTrades가 없으므로 0
				TakerBuyBaseVolume:  utility.ParseToFloat64(takerBuyBaseVolume),  // 바이낸스에는 TakerBuyBaseVolume이 없음
				TakerBuyQuoteVolume: utility.ParseToFloat64(takerBuyQuoteVolume), // 바이낸스에는 TakerBuyQuoteVolume이 없음
			})
		}
	} else if ReqInfo.ExchgNm == "bitget" {
		// 비트겟 응답 처리
		var rawData struct {
			Code        string          `json:"code"`
			Msg         string          `json:"msg"`
			RequestTime int64           `json:"requestTime"`
			Data        [][]interface{} `json:"data"`
		}

		if err := json.Unmarshal(body, &rawData); err != nil {
			return false, fmt.Sprintf("Error unmarshaling Bitget response: %v", err), nil
		}

		// rawData를 KlineResp 구조체로 변환
		for _, kline := range rawData.Data {
			if len(kline) < 6 {
				// 최소 6개의 필드가 있어야 함
				continue
			}

			// 공통 필드들 (비트겟은 timestamp가 string으로 반환됨)
			timestampStr := kline[0].(string)
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64) // string을 int64로 변환
			if err != nil {
				fmt.Println("Error parsing timestamp:", err)
				continue
			}
			open := kline[1].(string)
			high := kline[2].(string)
			low := kline[3].(string)
			close := kline[4].(string)
			volume := kline[5].(string)

			// KlineResp 구조체에 데이터를 할당
			resp = append(resp, bsml.KlineResp{
				Timestamp:           timestamp,
				Open:                utility.ParseToFloat64(open),
				High:                utility.ParseToFloat64(high),
				Low:                 utility.ParseToFloat64(low),
				Close:               utility.ParseToFloat64(close),
				Volume:              utility.ParseToFloat64(volume),
				CloseTime:           0,   // 비트겟에는 CloseTime이 없음
				QuoteVolume:         0.0, // 비트겟에서는 quoteVolume을 사용할 수 없으므로 0으로 처리
				NumberOfTrades:      0,   // 비트겟에는 NumberOfTrades가 없음
				TakerBuyBaseVolume:  0.0, // 비트겟에는 TakerBuyBaseVolume이 없음
				TakerBuyQuoteVolume: 0.0, // 비트겟에는 TakerBuyQuoteVolume이 없음
			})
		}
	}

	return true, "", resp
}

// 	// 2D 배열로 된 JSON 응답 파싱
// 	var rawData [][]interface{} // JSON을 먼저 2D 배열로 Unmarshal
// 	if err := json.Unmarshal(body, &rawData); err != nil {
// 		return false, fmt.Sprintf("Error unmarshaling response: %v", err), nil
// 	}

// 	// rawData를 KlineResp 구조체로 변환
// 	for _, kline := range rawData {
// 		if len(kline) < 6 {
// 			// 최소 6개의 필드가 있어야 함(bitget의 경우)
// 			continue
// 		}

// 		// 공통 필드들
// 		timestamp := int64(kline[0].(float64))
// 		open := kline[1].(string)
// 		high := kline[2].(string)
// 		low := kline[3].(string)
// 		close := kline[4].(string)
// 		baseVolume := kline[5].(string)

// 		var closeTime int64
// 		var numberOfTrades int
// 		var takerBuyBaseVolume, takerBuyQuoteVolume float64

// 		// 바이낸스일 경우 추가 필드 처리 (바이낸스는 12개 필드 제공)
// 		if ReqInfo.ExchgNm == "binance" && len(kline) >= 12 {
// 			closeTime = int64(kline[6].(float64))
// 			numberOfTrades = int(kline[8].(float64))
// 			takerBuyBaseVolume = utility.ParseToFloat64(kline[9].(string))
// 			takerBuyQuoteVolume = utility.ParseToFloat64(kline[10].(string))
// 		}

// 		// KlineResp 구조체에 데이터를 할당
// 		resp = append(resp, bsml.KlineResp{
// 			Timestamp:           timestamp,
// 			Open:                utility.ParseToFloat64(open),
// 			High:                utility.ParseToFloat64(high),
// 			Low:                 utility.ParseToFloat64(low),
// 			Close:               utility.ParseToFloat64(close),
// 			BaseVolume:          utility.ParseToFloat64(baseVolume),
// 			QuoteVolume:         0.0, // 비트겟에서는 quoteVolume을 사용할 수 없으므로 0으로 처리
// 			CloseTime:           closeTime,
// 			NumberOfTrades:      numberOfTrades,
// 			TakerBuyBaseVolume:  takerBuyBaseVolume,
// 			TakerBuyQuoteVolume: takerBuyQuoteVolume,
// 		})
// 	}

// 	return true, "", resp
// }
