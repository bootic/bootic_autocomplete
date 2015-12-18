package lib

import (
	"github.com/leekchan/accounting"
	"strconv"
)

func PageValue(rawValue string, defValue uint64) (val uint64) {
	if rawValue != "" {
		pval, err := strconv.ParseUint(rawValue, 10, 64)
		if err == nil {
			val = pval
		} else {
			val = defValue
		}
	} else {
		val = defValue
	}

	return val
}

var currencies = map[string]accounting.Accounting{
	"CLP": accounting.Accounting{Symbol: "$", Precision: 0, Thousand: ".", Decimal: ","},
}

func formatAmount(amount uint64, currency string) string {
	if ac, ok := currencies[currency]; ok {
		return ac.FormatMoney(amount)
	} else {
		return strconv.FormatUint(amount, 10)
	}
}
