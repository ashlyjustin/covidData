package handlers

import (
	"fmt"
	"time"
)

var StateCodesList = [...]string{
	"AN",
	"AP",
	"AR",
	"AS",
	"BR",
	"CH",
	"CT",
	"DL",
	"DN",
	"GA",
	"GJ",
	"HP",
	"HR",
	"JH",
	"JK",
	"KA",
	"KL",
	"LA",
	"LD",
	"MH",
	"ML",
	"MN",
	"MP",
	"MZ",
	"NL",
	"OR",
	"PB",
	"PY",
	"RJ",
	"SK",
	"TG",
	"TN",
	"TR",
	"TT",
	"UP",
	"UT",
	"WB",
}

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
	Confirmed int `json:"confirmed" bson:"confirmed"`
	Tested    int `json:"tested" bson:"tested"`
	Recovered int `json:"recovered" bson:"recovered"`
	Deceased  int `json:"deceased" bson:"deceased"`
}
type MetaData struct {
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
	Population  int       `json:"population" bson:"population"`
}
type User struct {
	Ip    string `json:"ip"`
	State State  `json:"state"`
}
type ErrorMessage struct {
	Message string `json:"message"`
}
