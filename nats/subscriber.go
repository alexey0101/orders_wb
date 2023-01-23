package nats

import (
	"app/db/impl"
	"app/model"
	"encoding/json"
	"errors"
	"log"

	stan "github.com/nats-io/stan.go"
)

func Subscribe(cancelChan <-chan struct{}, cache *impl.Cache, clusterId, clientId, subject, qgroup string) {
	sc, err := stan.Connect(clusterId, clientId)
	if err != nil {
		return
	}
	defer sc.Close()

	sc.QueueSubscribe(subject, qgroup, func(msg *stan.Msg) {
		order, err := messageToOrder(msg)
		if err != nil {
			log.Default().Println("Couldn't read published order: ", err)
			return
		}

		err = cache.SaveOrder(order)
		if err != nil {
			log.Default().Println("Couldn't save order: ", err)
			return
		}

		log.Default().Printf("Order %s saved successfully!", order.Order_uid)
	})

	<-cancelChan
}

func messageToOrder(msg *stan.Msg) (model.Order, error) {
	if !json.Valid(msg.Data) {
		return model.Order{}, errors.New("Order json file is not valid!")
	}

	var order model.Order
	err := json.Unmarshal(msg.Data, &order)

	return order, err
}
