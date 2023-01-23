package api

import (
	"app/model"
)

type OrderRepositoryInterface interface {
	Save(order *model.Order) error
	FindAll() ([]model.Order, error)
	Close() error
}
