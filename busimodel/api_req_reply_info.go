package busimodel

// KlineRequest : Kline API 요청의 파라미터들을 담는 구조체
type KlineRequest struct {
	ExchgNm   string `json:"exchange_name" binding:"required"` // 거래소 명
	Symbol    string `json:"symbol" binding:"required"`        // 거래쌍 (ex: BTC-USDT)
	Interval  string `json:"interval" binding:"required"`      // 간격 (ex: 1m, 5m, 1h...)
	StartTime string `json:"start_time" binding:"omitempty"`   // 시작 시간
	EndTime   string `json:"end_time" binding:"omitempty"`     // 종료 시간
	TimeZone  string `json:"time_zone" binding:"omitempty"`    // Default: 0 (UTC)
	Limit     int64  `json:"limit" binding:"omitempty"`        // Default 500; max 1000.
}

type KlineResp struct {
	Timestamp           int64   `json:"timestamp"`                        // UNIX timestamp
	Open                float64 `json:"open"`                             // Open price
	High                float64 `json:"high"`                             // High price
	Low                 float64 `json:"low"`                              // Low price
	Close               float64 `json:"close"`                            // Close price
	Volume              float64 `json:"volume"`                           // Volume in base currency (e.g., BTC in BTCUSDT)
	CloseTime           int64   `json:"close_time,omitempty"`             // Close time for Binance (may not be available for Bitget)
	QuoteVolume         float64 `json:"quote_volume"`                     // Volume in quote currency (e.g., USDT in BTCUSDT)
	NumberOfTrades      int     `json:"number_of_trades,omitempty"`       // Number of trades (for Binance)
	TakerBuyBaseVolume  float64 `json:"taker_buy_base_volume,omitempty"`  // Taker buy volume (for Binance)
	TakerBuyQuoteVolume float64 `json:"taker_buy_quote_volume,omitempty"` // Taker buy volume in quote currency (for Binance)
}
