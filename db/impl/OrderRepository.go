package impl

import (
	"app/db/api"
	"app/model"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type orderRepository struct {
	db         *sql.DB
	statements map[string]*sql.Stmt
	api.OrderRepositoryInterface
}

const (
	DB_DRIVER   = "postgres"
	DB_HOST     = "localhost"
	DB_PORT     = 5432
	DB_NAME     = "WB_DB"
	DB_USERNAME = ""
	DB_PASSWORD = ""

	deliveryInsertQuery = `
	insert into delivery(name, phone, zip, city, address, region, email)
	values ($1, $2, $3, $4, $5, $6, $7) returning delivery_id`
	itemInsertQuery = `
	insert into item(chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning item_id`
	paymentInsertQuery = `
	insert into payment(transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning transaction_id`
	orderInsertQuery = `
	insert into "order"(order_uid, track_number, entry, delivery_id, transaction_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning order_id`
	orderItemInsertQuery = `
	insert into orderitem(order_id, item_id)
	values ($1, $2) returning orderitem_id`
	findPaymentByIdQuery = `
	select * from payment where transaction_id = $1`
	findDeliveryByIdQuery = `
	select * from delivery where delivery_id = $1`
	findItemsByOrderIdQuery = `
	select item.item_id, chrt_id, item.track_number, price, rid, item.name, sale, size, total_price, nm_id, brand, status from item
	join orderitem on
	orderitem.item_id = item.item_id
	join "order" on
	"order".order_id = orderitem.order_id
	where "order".order_id = $1`
	findAllOrdersQuery = `select order_id, track_number, order_uid, entry, delivery_id, transaction_id, locale, internal_signature,
	customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard from "order"`
)

func NewOrderRepository() (api.OrderRepositoryInterface, error) {
	dataSourceInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_NAME)
	db, err := sql.Open(DB_DRIVER, dataSourceInfo)
	if err != nil {
		return nil, err
	}

	statements, err := getSaveStatements(db)
	if err != nil {
		return nil, err
	}

	return &orderRepository{db: db, statements: statements}, err
}

func getSaveStatements(db *sql.DB) (map[string]*sql.Stmt, error) {
	statements := map[string]*sql.Stmt{}

	deliveryInsertStmt, err := db.Prepare(deliveryInsertQuery)
	if err != nil {
		return nil, err
	}
	statements[deliveryInsertQuery] = deliveryInsertStmt

	itemInsertStmt, err := db.Prepare(itemInsertQuery)
	if err != nil {
		return nil, err
	}
	statements[itemInsertQuery] = itemInsertStmt

	paymentInsertStmt, err := db.Prepare(paymentInsertQuery)
	if err != nil {
		return nil, err
	}
	statements[paymentInsertQuery] = paymentInsertStmt

	orderInsertStmt, err := db.Prepare(orderInsertQuery)
	if err != nil {
		return nil, err
	}
	statements[orderInsertQuery] = orderInsertStmt

	orderItemInsertStmt, err := db.Prepare(orderItemInsertQuery)
	if err != nil {
		return nil, err
	}
	statements[orderItemInsertQuery] = orderItemInsertStmt

	findDeliveryByIdStmt, err := db.Prepare(findDeliveryByIdQuery)
	if err != nil {
		return nil, err
	}
	statements[findDeliveryByIdQuery] = findDeliveryByIdStmt

	findPaymentByIdStmt, err := db.Prepare(findPaymentByIdQuery)
	if err != nil {
		return nil, err
	}
	statements[findPaymentByIdQuery] = findPaymentByIdStmt

	findItemsByOrderIdStmt, err := db.Prepare(findItemsByOrderIdQuery)
	if err != nil {
		return nil, err
	}
	statements[findItemsByOrderIdQuery] = findItemsByOrderIdStmt

	findAllOrdersStmt, err := db.Prepare(findAllOrdersQuery)
	if err != nil {
		return nil, err
	}
	statements[findAllOrdersQuery] = findAllOrdersStmt

	return statements, nil
}

func (orderRep *orderRepository) Save(order *model.Order) error {
	tx, err := orderRep.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deliveryId, err := orderRep.insertRow(tx.Stmt(orderRep.statements[deliveryInsertQuery]), order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}
	order.Delivery.Delivery_id = deliveryId

	transaction_id, err := orderRep.insertRow(tx.Stmt(orderRep.statements[paymentInsertQuery]), order.Payment.Transaction, order.Payment.Request_id, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.Payment_dt, order.Payment.Bank, order.Payment.Delivery_cost, order.Payment.Goods_total, order.Payment.Custom_fee)
	if err != nil {
		return err
	}
	order.Payment.Transaction_id = transaction_id

	for idx := range order.Items {
		item_id, err := orderRep.insertRow(tx.Stmt(orderRep.statements[itemInsertQuery]), order.Items[idx].Chrt_id, order.Items[idx].Track_number, order.Items[idx].Price, order.Items[idx].Rid,
			order.Items[idx].Name, order.Items[idx].Sale, order.Items[idx].Size, order.Items[idx].Total_price, order.Items[idx].Nm_id, order.Items[idx].Brand, order.Items[idx].Status)
		if err != nil {
			return err
		}
		order.Items[idx].Item_id = item_id
	}

	order_id, err := orderRep.insertRow(tx.Stmt(orderRep.statements[orderInsertQuery]), order.Order_uid, order.Track_number, order.Entry, order.Delivery.Delivery_id,
		order.Payment.Transaction_id, order.Locale, order.Internal_signature, order.Customer_id,
		order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created, order.Oof_shard)
	if err != nil {
		return err
	}
	order.Order_id = order_id

	for _, val := range order.Items {
		_, err := orderRep.insertRow(tx.Stmt(orderRep.statements[orderItemInsertQuery]), order.Order_id, val.Item_id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (orderRep *orderRepository) insertRow(statement *sql.Stmt, params ...any) (int64, error) {
	var id int64
	err := statement.QueryRow(params...).Scan(&id)

	return id, err
}

func (orderRep *orderRepository) FindAll() ([]model.Order, error) {
	tx, err := orderRep.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Stmt(orderRep.statements[findAllOrdersQuery]).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}

	for rows.Next() {
		var order model.Order
		rows.Scan(&order.Order_id, &order.Track_number, &order.Order_uid, &order.Entry, &order.Delivery.Delivery_id, &order.Payment.Transaction_id, &order.Locale,
			&order.Internal_signature, &order.Customer_id, &order.Delivery_service, &order.Shardkey,
			&order.Sm_id, &order.Date_created, &order.Oof_shard)
		err := orderRep.statements[findDeliveryByIdQuery].QueryRow(order.Delivery.Delivery_id).Scan(&order.Delivery.Delivery_id, &order.Delivery.Name, &order.Delivery.Phone,
			&order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
		if err == sql.ErrNoRows {
			return nil, err
		}
		err = orderRep.statements[findPaymentByIdQuery].QueryRow(order.Payment.Transaction_id).Scan(&order.Payment.Transaction_id, &order.Payment.Transaction, &order.Payment.Request_id, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
			&order.Payment.Payment_dt, &order.Payment.Bank, &order.Payment.Delivery_cost, &order.Payment.Goods_total, &order.Payment.Custom_fee)
		if err == sql.ErrNoRows {
			return nil, err
		}
		orders = append(orders, order)
	}

	for idx := range orders {
		rows, err := tx.Stmt(orderRep.statements[findItemsByOrderIdQuery]).Query(orders[idx].Order_id)
		if err != nil {
			return nil, err
		}

		items := []model.Item{}
		for rows.Next() {
			var item model.Item
			rows.Scan(&item.Item_id, &item.Chrt_id, &item.Track_number, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
				&item.Total_price, &item.Nm_id, &item.Brand, &item.Status)
			items = append(items, item)
		}
		rows.Close()
		orders[idx].Items = items
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (orderRep *orderRepository) Close() error {
	return orderRep.db.Close()
}
