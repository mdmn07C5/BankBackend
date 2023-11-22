package util

const (
	USD = "USD"
	CAD = "CAD"
	EUR = "EUR"
	MXN = "MXN"
	GBP = "GBP"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, CAD, EUR, MXN, GBP:
		return true
	}
	return false
}
