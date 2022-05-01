package main

import (
	"context"
	"covidApp/handlers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var e = echo.New()
var mongoClient mongo.Client
var (
	c        *mongo.Client
	db       *mongo.Database
	stateCol *mongo.Collection
)

func init() {
	fmt.Println("inside init")
	err := cleanenv.ReadConfig("config.yml", &handlers.Cfg)
	if err != nil {
		e.Logger.Fatal("Unable to load configuration")
	}
	fmt.Println(handlers.Cfg)
	createMongoConnection()
}
func createMongoConnection() {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://ashlyjustin:iamgroot@coviddatacluster.ssyqx.mongodb.net/Covid?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	fmt.Println("connected")
	defer cancel()
	var err error
	c, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	db = c.Database(handlers.Cfg.Database)
	stateCol = db.Collection(handlers.Cfg.Collection)
	fmt.Println(c, stateCol, "state collection")
	// postCovidData(c)

}
func postCovidData() {
	fmt.Println("inside mongo client post")

	var allStateData []handlers.State
	allStateData = getAllCovidData()
	fmt.Println("reeached insert", allStateData)
	count := 0
	for _, state := range allStateData {
		filter := bson.M{"StateCode": state.StateCode}
		// not:=bson.M{"Total":state.Total,"Meta":state.Meta}}
		update := bson.M{"$set": state}
		opts := options.Update().SetUpsert(true)

		_, err := stateCol.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			panic(err)
		}
		count++
	}
	fmt.Println("Total updated objects : ", count)

	// covidCollection.InsertMany(json.Marshal(all÷StateData))
}
func getAllCovidData() []handlers.State {
	fmt.Println("inside get covid data")
	client := &http.Client{}
	desc, err := client.Get(handlers.Cfg.CovidDataUrl)
	if err != nil {
		fmt.Print(err)
	}
	jsonData, err := ioutil.ReadAll(desc.Body)
	if err != nil {
		panic(err)
	}
	allStateData := make(map[string]json.RawMessage)
	// unmarschal JSON
	e := json.Unmarshal(jsonData, &allStateData)

	if e != nil {
		panic(e)
	}

	var stateData []handlers.State
	for key, value := range allStateData {
		var state handlers.State
		e := json.Unmarshal(value, &state)
		if e != nil {
			panic(e)
		}
		state.StateCode = key
		stateData = append(stateData, state)
	}
	return stateData
}
func main() {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	h := &handlers.StateHandler{Col: stateCol}
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/getStateData", h.GetStateData)
	e.Logger.Fatal(e.Start(":1323"))
}
