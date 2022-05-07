package util

// constants for all supported currencies
const (
	USD = "USD"
	INR = "INR"
	CAD = "CAD"
	EUR = "EUR"
)

// IsSupportedCurrency return true if input currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, INR, CAD, EUR:
		return true
	}
	return false
}
