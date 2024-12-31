package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	bsml "krononlabs/busimodel"
	utility "krononlabs/utility"

	"github.com/gorilla/websocket"
)

const (
	binanceKline = "https://api.binance.com/api/v3/klines"
	bitgetKline  = "https://api.bitget.com/api/v2/spot/market/history-candles"

	binanWSKline = ""
	bitgtWSKline = ""
)

// Kline 통합 조회
func KlineInq(ReqInfo *bsml.KlineRequest) (bool, string, map[string][]bsml.KlineResp) {

	var (
		respMap = make(map[string][]bsml.KlineResp) // 심볼별로 응답을 저장하는 맵
		wg      sync.WaitGroup                      // 동시 요청을 처리하기 위한 WaitGroup
		mu      sync.Mutex                          // 여러 고루틴에서 공유하는 resp를 안전하게 수정하기 위한 Mutex
	)

	// 거래소별 interval 값 변환
	interval := utility.GetExchangeInterval(ReqInfo.Interval, ReqInfo.ExchgNm)

	// Symbols 배열에 대해 각각의 요청을 병렬 처리
	for _, symbol := range ReqInfo.Symbols { // 이제 Symbols는 배열로 가정
		wg.Add(1)
		go func(symbolStr string) {
			defer wg.Done()

			var url string

			// 거래소별 Kline API 요청 URL 생성
			if ReqInfo.ExchgNm == "binance" {
				// 바이낸스 Kline API 요청 URL 생성
				url = fmt.Sprintf("%s?symbol=%s&interval=%s", binanceKline, symbolStr, interval)
			} else if ReqInfo.ExchgNm == "bitget" {
				// 비트겟 Kline API 요청 URL 생성
				url = fmt.Sprintf("%s?symbol=%s&granularity=%s", bitgetKline, symbolStr, interval)
			}

			// 선택적인 값이 존재하면 URL에 추가
			if ReqInfo.StartTime != "" {
				// 날짜 문자열을 time.Time 객체로 변환
				start, err := time.Parse(time.RFC3339, ReqInfo.StartTime)
				if err != nil {
					fmt.Println("Error parsing start time:", err)
					return
				}
				startUnix := start.Unix() * 1000

				url = fmt.Sprintf("%s&startTime=%d", url, startUnix)
			}

			if ReqInfo.EndTime != "" {
				end, err := time.Parse(time.RFC3339, ReqInfo.EndTime)
				if err != nil {
					fmt.Println("Error parsing end time:", err)
					return
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

			fmt.Println(url)

			// HTTP GET 요청 보내기
			httpResp, err := http.Get(url)
			if err != nil {
				fmt.Println("Request failed for symbol:", symbol)
				return
			}
			fmt.Println(httpResp)
			defer httpResp.Body.Close()

			body, err := io.ReadAll(httpResp.Body)
			fmt.Println(body)
			if err != nil {
				fmt.Println("Error reading response body for symbol:", symbol)
				return
			}

			// 거래소에 따라 응답을 다르게 처리
			if ReqInfo.ExchgNm == "binance" {
				// 바이낸스 응답 처리
				var rawData [][]interface{}
				if err := json.Unmarshal(body, &rawData); err != nil {
					fmt.Println("Error unmarshaling Binance response:", err)
					return
				}

				// rawData를 KlineResp 구조체로 변환
				var respData []bsml.KlineResp
				for _, kline := range rawData {
					if len(kline) < 11 {
						// 최소 11개의 필드가 있어야 함
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
					respData = append(respData, bsml.KlineResp{
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
				// 심볼별로 응답을 맵에 추가
				mu.Lock()
				respMap[symbolStr] = respData
				mu.Unlock()
			} else if ReqInfo.ExchgNm == "bitget" {
				// 비트겟 응답 처리
				var rawData struct {
					Code        string          `json:"code"`
					Msg         string          `json:"msg"`
					RequestTime int64           `json:"requestTime"`
					Data        [][]interface{} `json:"data"`
				}

				if err := json.Unmarshal(body, &rawData); err != nil {
					fmt.Println("Error unmarshaling Bitget response:", err)
					return
				}

				// rawData를 KlineResp 구조체로 변환
				var respData []bsml.KlineResp
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
					mu.Lock() // 응답을 안전하게 추가하기 위해 락을 사용
					respData = append(respData, bsml.KlineResp{
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
				// 심볼별로 응답을 맵에 추가
				mu.Lock()
				respMap[symbolStr] = respData
				mu.Unlock()
			}
		}(symbol.Symbol) // 고루틴에 symbol을 전달
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	// 결과 반환
	return true, "", respMap
}

// KlineWebSocoketInq
func KlineWebSocoketInqBinance(ReqInfo *bsml.WebSocketKlineRequest) (bool, string, bsml.WebSocketResponse) {

	var resp bsml.WebSocketResponse

	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_%s", ReqInfo.Symbol, ReqInfo.Interval)

	// 웹소켓 연결
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return false, "websocket connect failed", resp
	}
	defer conn.Close()

	fmt.Println("웹소켓 연결 성공")

	// 웹소켓으로부터 메시지를 지속적으로 수신
	for {
		// 메시지 수신
		_, message, err := conn.ReadMessage()
		if err != nil {
			return false, "", resp
		}

		// 받은 원시 JSON 데이터를 그대로 Python 스크립트로 전달
		cmd := exec.Command("python3", "multiAssetCryptoStrategy.py")

		// 원시 JSON 데이터를 Python 스크립트에 전달
		cmd.Stdin = bytes.NewReader(message) // 받은 JSON 메시지를 stdin으로 전달

		// Python 스크립트 실행 결과 받기
		output, err := cmd.CombinedOutput()
		if err != nil {
			return false, fmt.Sprintf("error executing python script: %s", err.Error()), resp
		}

		// Python 스크립트의 결과 출력
		fmt.Printf("Python script output: %s\n", output)

		// 로그로 받은 원시 데이터 출력 (디버깅용)
		fmt.Printf("Received raw Kline data: %s\n", string(message))

		// // JSON 파싱
		// err = json.Unmarshal(message, &resp)
		// fmt.Println(err)
		// if err != nil {
		// 	return false, err.Error(), resp
		// 	continue
		// }

		// // Kline 데이터 출력
		// fmt.Printf("Timestamp: %d, Open: %s, Close: %s, High: %s, Low: %s, Volume: %s\n",
		// 	resp.Kline.Timestamp, resp.Kline.OpenPrice, resp.Kline.ClosePrice,
		// 	resp.Kline.HighPrice, resp.Kline.LowPrice, resp.Kline.Volume)
	}
}

// 구독 메시지 구조체
type SubscribeMessage struct {
	Op   string `json:"op"`
	Args []struct {
		InstType string `json:"instType"`
		Channel  string `json:"channel"`
		InstId   string `json:"instId"`
	} `json:"args"`
}

// WebSocket 연결 및 Kline 데이터 구독 함수
func KlineWebSocoketInqBitget(ReqInfo *bsml.WebSocketKlineRequest) (bool, string, bsml.WebSocketResponse) {

	var resp bsml.WebSocketResponse

	const url = "wss://ws.bitget.com/v2/ws/public"
	// WebSocket 연결
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return false, "websocket connect failed", resp
	}
	defer conn.Close()

	// Kline 데이터 및 Ticker 구독 메시지 생성
	subscribeMessage := SubscribeMessage{
		Op: "subscribe",
		Args: []struct {
			InstType string `json:"instType"`
			Channel  string `json:"channel"`
			InstId   string `json:"instId"`
		}{
			{
				InstType: "SPOT",                      // 상품 종류
				Channel:  "candle" + ReqInfo.Interval, // 예: "candle5m"으로 Kline 채널 설정
				InstId:   ReqInfo.Symbol,              // 거래쌍, 예: "BTC-USDT"
			},
		},
	}

	// 구독 메시지 JSON으로 변환
	subscribeMessageJSON, err := json.Marshal(subscribeMessage)
	if err != nil {
		return false, "websocket connect failed", resp
	}

	fmt.Println(subscribeMessageJSON)

	// 구독 메시지 전송
	err = conn.WriteMessage(websocket.TextMessage, subscribeMessageJSON)
	if err != nil {
		log.Fatal("구독 메시지 전송 실패:", err)
	}
	fmt.Println("BTCUSDT 5분 간격 Kline 데이터 구독 시작")

	// 실시간 데이터 수신
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("메시지 수신 오류:", err)
		}

		// 수신된 메시지 출력 (그대로 출력하여 확인)
		fmt.Printf("수신된 메시지: %s\n", message)
	}
}

// 통합된 WebSocket 요청 처리 함수
func KlineWebSocoketInq(ReqInfo *bsml.WebSocketKlineRequest) (bool, string, bsml.WebSocketResponse) {
	if ReqInfo.ExchgNm == "binance" {

		resultFlag, errMsg, resp := KlineWebSocoketInqBinance(ReqInfo)

		return resultFlag, errMsg, resp
	} else if ReqInfo.ExchgNm == "bitget" {

		resultFlag, errMsg, resp := KlineWebSocoketInqBitget(ReqInfo)

		return resultFlag, errMsg, resp
	} else {
		return false, "지원되지 않는 거래소 입니다.", bsml.WebSocketResponse{}
	}
}

// // Go에서 받은 원시 JSON 데이터를 파이썬 스크립트로 전달하는 함수
// func sendToPython(rawData []byte) (string, error) {
// 	// 파이썬 스크립트 경로
// 	pythonScript := "multiAssetCryptoStrategy.py"

// 	// 파이썬 스크립트 실행
// 	cmd := exec.Command("python", pythonScript)
// 	cmd.Stdin = bytes.NewReader(rawData)

// 	// 파이썬의 표준 출력 결과를 캡처할 변수
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &out

// 	// 파이썬 스크립트 실행
// 	err := cmd.Run()
// 	if err != nil {
// 		return "", fmt.Errorf("python script execution failed: %s", err)
// 	}

// 	fmt.Println(out.String())

// 	// 파이썬에서 반환한 결과를 반환
// 	return out.String(), nil
// }

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
