package domain

type Coin struct {
	Id     int64  `db:"id" json:"id"`
	Name   string `db:"coin_name" json:"name"`
	Symbol string `db:"symbol" json:"symbol"`
}
