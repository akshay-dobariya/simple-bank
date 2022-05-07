package api

import (
	"github.com/akshay-dobariya/simple-bank/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(filedLevel validator.FieldLevel) bool {
	if currency, ok := filedLevel.Field().Interface().(string); ok {
		// check if currency is supported or not
		return util.IsSupportedCurrency(currency)
	}
	return false
}
