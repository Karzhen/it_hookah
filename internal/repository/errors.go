package repository

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInsufficientStock = errors.New("insufficient stock")
var ErrEmptyCart = errors.New("empty cart")
var ErrForbidden = errors.New("forbidden")
var ErrInvalidStockOperation = errors.New("invalid stock operation")
var ErrInvalidOrderStatus = errors.New("invalid order status")
var ErrInvalidMixPercentTotal = errors.New("invalid mix percent total")
var ErrInvalidMixItemProduct = errors.New("invalid mix item product")
var ErrTobaccoProductNotFound = errors.New("tobacco product not found")
var ErrDuplicateTagCode = errors.New("duplicate tag code")
var ErrDuplicateTagName = errors.New("duplicate tag name")
