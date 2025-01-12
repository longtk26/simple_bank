package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/longtk26/simple_bank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// check if currency is valid
		return util.IsSupportedCurrency(currency)
	}

	return false
}