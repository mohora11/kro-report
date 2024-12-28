package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	bsml "krononlabs/busimodel"
)

const (
	binanceKline = "https://api.binance.com/api/v1/klines"
	bitgetKline  = "https://api.bitget.com/api/v1/market/candles"
)

// Kline 조회
func KlineInq(ReqInfo *bsml.KlineRequest) (bool, string, bsml.KlineResp) {
	url := fmt.Sprintf("%s?symbol=%s&interval=%s&startTime=%s&endTime=%s&timeZone=%s&limit=%s", binanceKline, ReqInfo.Symbol, ReqInfo.Interval)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines []KlineResp
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, err
	}

	return klines, nil
}
