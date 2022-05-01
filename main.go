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

func init() {
	fmt.Println("inside init")
	err := cleanenv.ReadConfig("config.yml", &handlers.Cfg)
	if err != nil {
		e.Logger.Fatal("Unable to load configuration")
	}
	createMongoConnection()
}
func createMongoConnection() {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://ashlyjustin:iamgroot@coviddatacluster.ssyqx.mongodb.net/Covid?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	fmt.Println("connected")
	defer cancel()
	var err error
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	covidDatabase := mongoClient.Database("Covid")
	covidCollection := covidDatabase.Collection("StateData")
	var allStateData []handlers.State
	allStateData = getAllCovidData()
	fmt.Println("reeached insert", allStateData)
	for _, state := range allStateData {
		filter := bson.M{"StateCode": state.StateCode}
		// not:=bson.M{"Total":state.Total,"Meta":state.Meta}}
		update := bson.M{"$set": state}
		opts := options.Update().SetUpsert(true)

		result, err := covidCollection.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
		fmt.Printf("Number of documents updated: %v\n", result.ModifiedCount)
		fmt.Printf("Number of documents upserted: %v\n", result.UpsertedCount)
	}

	fmt.Println(mongoClient)
	// postCovidData()

}
func postCovidData() {

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
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
