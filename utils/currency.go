package utils

import (
	"fmt"
)

// Configuration des taux de change par défaut
const (
	DefaultUSDToCDF = 2700.0  // Taux par défaut USD vers CDF
	DefaultCDFToUSD = 0.00037 // Taux par défaut CDF vers USD
)

// ConvertCurrency convertit un montant d'une devise à une autre en utilisant un taux manuel
func ConvertCurrency(amount float64, rate float64) float64 {
	return amount * rate
}

// GetDefaultExchangeRate retourne les taux de change par défaut
func GetDefaultExchangeRate(from, to string) float64 {
	switch {
	case from == "USD" && to == "CDF":
		return DefaultUSDToCDF
	case from == "CDF" && to == "USD":
		return DefaultCDFToUSD
	default:
		return 1.0 // Aucune conversion par défaut
	}
}

// ConvertUSDToCDF convertit USD vers CDF avec un taux donné
func ConvertUSDToCDF(usdAmount float64, rate float64) float64 {
	if rate <= 0 {
		rate = DefaultUSDToCDF
	}
	return ConvertCurrency(usdAmount, rate)
}

// ConvertCDFToUSD convertit CDF vers USD avec un taux donné
func ConvertCDFToUSD(cdfAmount float64, rate float64) float64 {
	if rate <= 0 {
		rate = DefaultCDFToUSD
	}
	return ConvertCurrency(cdfAmount, rate)
}

// ConvertWithDefaultRate convertit en utilisant les taux par défaut
func ConvertWithDefaultRate(amount float64, fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	rate := GetDefaultExchangeRate(fromCurrency, toCurrency)
	if rate == 1.0 && fromCurrency != toCurrency {
		return 0, fmt.Errorf("no conversion rate available for %s to %s", fromCurrency, toCurrency)
	}

	return ConvertCurrency(amount, rate), nil
}
