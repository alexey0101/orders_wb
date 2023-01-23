package impl

import (
	"app/db/api"
	"app/model"
	"errors"
	"sync"
)

type Cache struct {
	Orders          sync.Map
	orderRepository api.OrderRepositoryInterface
}

func NewCache() (*Cache, error) {
	orders := sync.Map{}
	orderRepository, err := NewOrderRepository()
	if err != nil {
		return nil, err
	}

	return &Cache{orders, orderRepository}, nil
}

func (cache *Cache) LoadOrders() error {
	orders, err := cache.orderRepository.FindAll()
	if err != nil {
		return err
	}
	for _, order := range orders {
		cache.Orders.Store(order.Order_uid, order)
	}

	return nil
}

func (cache *Cache) SaveOrder(order model.Order) error {
	err := cache.orderRepository.Save(&order)
	if err != nil {
		return err
	}
	cache.Orders.Store(order.Order_uid, order)
	return nil
}

func (cache *Cache) FindByUid(uid string) (model.Order, error) {
	order, ok := cache.Orders.Load(uid)

	if !ok {
		return model.Order{}, errors.New("There is no order with given uid!")
	}

	return order.(model.Order), nil
}
