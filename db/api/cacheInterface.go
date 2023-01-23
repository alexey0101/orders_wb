package api

import "app/model"

type Cache interface {
	LoadOrders() error
	SaveOrder(order model.Order) error
	FindByUid(uid string) (model.Order, error)
}
