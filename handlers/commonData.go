package handlers

import (
	"fmt"
	"time"
)

func Common() {
	fmt.Println("from common")
}

type State struct {
	StateCode string    `json:"stateCode"`
	Total     TotalData `json:"total"`
	Meta      MetaData  `json:"meta"`
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
