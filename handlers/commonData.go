package handlers

import (
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
var StateNameMap = map[string]string{
	"AN": "Andaman & Nicobar Islands",
	"AP": "Andhra Pradesh",
	"AR": "Arunachal Pradesh",
	"AS": "Assam",
	"BR": "Bihar",
	"CH": "Chandigarh",
	"CT": "Chhattisgarh",
	"DL": "Delhi",
	"DN": "Dadra & Nagar Haveli",
	"GA": "Goa",
	"GJ": "Gujarat",
	"HP": "Himachal Pradesh",
	"HR": "Haryana",
	"JH": "Jharkhand",
	"JK": "Jammu & Kashmir",
	"KA": "Karnataka",
	"KL": "Kerala",
	"LA": "Ladakh",
	"LD": "Lakshwadeep",
	"MH": "Maharashtra",
	"ML": "Meghalaya",
	"MN": "Manipur",
	"MP": "Madhya Pradesh",
	"MZ": "Mizoram",
	"NL": "Nagaland",
	"OR": "Orrisa",
	"PB": "Punjab",
	"PY": "Pondicherry",
	"RJ": "Rajasthan",
	"SK": "Sikkim",
	"TG": "Telangana",
	"TN": "Tamil Nadu",
	"TR": "Tripura",
	"UP": "Uttar Pradesh",
	"UT": "Uttarakhand",
	"WB": "West Bengal",
}

type SingleState struct {
	StateName string `json:"StateName"`
	State     State  `json:"State" bson:"State"`
}

//add validations here of length
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

//add validations here of length  & convert to readable time form 24 hr format
type MetaData struct {
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
	Population  int       `json:"population" bson:"population"`
}
type UserState struct {
	Location GeoLocation `json:"location"`
	State    SingleState `json:"covidData"`
}
type ErrorMessage struct {
	Message string `json:"message"`
}
type GeoLocation struct {
	Region     string `json:"region" `
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	Timezone   string `json:"timezone"`
}
