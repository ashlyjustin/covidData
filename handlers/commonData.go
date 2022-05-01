package handlers

import (
	"fmt"
	"time"
)

func Common() {
	fmt.Println("from common")
}

type State struct {
	// ID        primitive.ObjectID `json:"_id" bson:"_id"`
	StateCode string    `json:"StateCode" bson:"StateCode"`
	Total     TotalData `json:"total" bson:"total"`
	Meta      MetaData  `json:"meta" bson:"meta"`
}
type TotalData struct {
	Confirmed int `json:"confirmed"`
	Tested    int `json:"tested"`
	Recovered int `json:"recovered"`
	Deceased  int `json:"deceased"`
}
type MetaData struct {
	LastUpdated time.Time `json:"last_updated"`
	Population  int       `json:"population"`
}
type User struct {
	ip string
}
type ErrorMessage struct {
	Message string `json:"message"`
}
