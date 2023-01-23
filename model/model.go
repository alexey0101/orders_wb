package model

type Order struct {
	Order_id           int64
	Order_uid          string
	Track_number       string
	Entry              string
	Delivery           Delivery
	Payment            Payment
	Items              []Item
	Locale             string
	Internal_signature string
	Customer_id        string
	Delivery_service   string
	Shardkey           string
	Sm_id              int
	Date_created       string
	Oof_shard          string
}

type Payment struct {
	Transaction_id int64
	Transaction    string
	Request_id     string
	Currency       string
	Provider       string
	Amount         int
	Payment_dt     int
	Bank           string
	Delivery_cost  int
	Goods_total    int
	Custom_fee     int
}

type Delivery struct {
	Delivery_id int64
	Name        string
	Phone       string
	Zip         string
	City        string
	Address     string
	Region      string
	Email       string
}

type Item struct {
	Item_id      int64
	Chrt_id      int
	Track_number string
	Price        int
	Rid          string
	Name         string
	Sale         int
	Size         string
	Total_price  int
	Nm_id        int
	Brand        string
	Status       int
}
