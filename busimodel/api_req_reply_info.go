package busimodel

import "time"

// KlineRequest : Kline API 요청의 파라미터들을 담는 구조체
type KlineRequest struct {
	Symbol    string    `json:"symbol" binding:"required"`      // 거래쌍 (ex: BTC-USDT)
	Interval  string    `json:"interval" binding:"required"`    // 간격 (ex: 1m, 5m, 1h...)
	StartTime time.Time `json:"start_time" binding:"omitempty"` // 시작 시간
	EndTime   time.Time `json:"end_time" binding:"omitempty"`   // 종료 시간
	TimeZone  string    `json:"time_zone" binding:"omitempty"`  // Default: 0 (UTC)
	Limit     int       `json:"limit" binding:"omitempty"`      // Default 500; max 1000.
}

// KlineRequest : Kline 회신 구조체
type KlineResp struct {
	Time       int64   `json:"time"`
	Open       float64 `json:"open"`
	Close      float64 `json:"close"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Volume     float64 `json:"volume"`
	CloseTime  int64   `json:"closeTime"`
	QuoteAsset float64 `json:"quoteAsset"`
}
