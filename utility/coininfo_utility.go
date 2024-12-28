package utility

import (
	"strconv"

	bizcnst "krononlabs/com"
)

// getBinanceInterval : 거래소별로 지원하는 interval을 변환
func getBinanceInterval(interval string) string {
	// 바이낸스에서는 그대로 사용
	switch interval {
	case "1s":
		return bizcnst.BinanceInterval1s
	case "1m":
		return bizcnst.BinanceInterval1m
	case "3m":
		return bizcnst.BinanceInterval3m
	case "5m":
		return bizcnst.BinanceInterval5m
	case "15m":
		return bizcnst.BinanceInterval15m
	case "30m":
		return bizcnst.BinanceInterval30m
	case "1h":
		return bizcnst.BinanceInterval1h
	case "2h":
		return bizcnst.BinanceInterval2h
	case "4h":
		return bizcnst.BinanceInterval4h
	case "6h":
		return bizcnst.BinanceInterval6h
	case "8h":
		return bizcnst.BinanceInterval8h
	case "12h":
		return bizcnst.BinanceInterval12h
	case "1d":
		return bizcnst.BinanceInterval1d
	case "3d":
		return bizcnst.BinanceInterval3d
	case "1w":
		return bizcnst.BinanceInterval1w
	case "1M":
		return bizcnst.BinanceInterval1M
	default:
		return bizcnst.BinanceInterval1d
	}
}

// getBitgetGranularity :
func getBitgetGranularity(interval string) string {
	// 비트겟에서는 granularity로 각 요청값이 차이가 있음
	switch interval {
	case "1m":
		return bizcnst.BitgetGranularity1m
	case "5m":
		return bizcnst.BitgetGranularity5m
	case "15m":
		return bizcnst.BitgetGranularity15m
	case "30m":
		return bizcnst.BitgetGranularity30m
	case "1h":
		return bizcnst.BitgetGranularity1h
	case "4h":
		return bizcnst.BitgetGranularity4h
	case "6h":
		return bizcnst.BitgetGranularity6h
	case "12h":
		return bizcnst.BitgetGranularity12h
	case "1d":
		return bizcnst.BitgetGranularity1day
	case "3d":
		return bizcnst.BitgetGranularity3day
	case "1w":
		return bizcnst.BitgetGranularity1w
	case "1M":
		return bizcnst.BitgetGranularity1M
	default:
		return bizcnst.BitgetGranularity1day
	}
}

// getExchangeInterval : 거래소별 interval에 맞게 변환
func GetExchangeInterval(interval, exchgNm string) string {

	switch exchgNm {
	case "binance":
		return getBinanceInterval(interval)
	case "bitget":
		return getBitgetGranularity(interval)
	default:
		return ""
	}
}

// 문자열을 float64로 변환하는 함수 (빈 문자열이 오면 0으로 처리)
func ParseToFloat64(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return result
}
