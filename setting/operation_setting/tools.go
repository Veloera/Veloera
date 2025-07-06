package operation_setting

import "strings"

const (
	// Web search
	WebSearchHighTierModelPriceLow    = 30.00
	WebSearchHighTierModelPriceMedium = 35.00
	WebSearchHighTierModelPriceHigh   = 50.00
	WebSearchPriceLow                 = 25.00
	WebSearchPriceMedium              = 27.50
	WebSearchPriceHigh                = 30.00
	// File search
	FileSearchPrice = 2.5
)

func GetWebSearchPricePerThousand(modelName string, contextSize string) float64 {
	// 确定模型类型
	// https://platform.openai.com/docs/pricing Web search 价格按模型类型和 search context size 收费
	// gpt-4.1, gpt-4o, or gpt-4o-search-preview 更贵，gpt-4.1-mini, gpt-4o-mini, gpt-4o-mini-search-preview 更便宜
	isHighTierModel := (strings.HasPrefix(modelName, "gpt-4.1") || strings.HasPrefix(modelName, "gpt-4o")) &&
		!strings.Contains(modelName, "mini")
	// 确定 search context size 对应的价格
	var priceWebSearchPerThousandCalls float64
	switch contextSize {
	case "low":
		if isHighTierModel {
			priceWebSearchPerThousandCalls = WebSearchHighTierModelPriceLow
		} else {
			priceWebSearchPerThousandCalls = WebSearchPriceLow
		}
	case "medium":
		if isHighTierModel {
			priceWebSearchPerThousandCalls = WebSearchHighTierModelPriceMedium
		} else {
			priceWebSearchPerThousandCalls = WebSearchPriceMedium
		}
	case "high":
		if isHighTierModel {
			priceWebSearchPerThousandCalls = WebSearchHighTierModelPriceHigh
		} else {
			priceWebSearchPerThousandCalls = WebSearchPriceHigh
		}
	default:
		// search context size 默认为 medium
		if isHighTierModel {
			priceWebSearchPerThousandCalls = WebSearchHighTierModelPriceMedium
		} else {
			priceWebSearchPerThousandCalls = WebSearchPriceMedium
		}
	}
	return priceWebSearchPerThousandCalls
}

func GetFileSearchPricePerThousand() float64 {
	return FileSearchPrice
}
