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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
	covidDatabase := client.Database("Covid")
	covidCollection := covidDatabase.Collection("StateData")
	podcastResult, err := covidCollection.InsertOne(ctx, bson.D{
		{Key: "title", Value: "The Polyglot Developer Podcast"},
		{Key: "author", Value: "Nic Raboy"},
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("hello", podcastResult)
	}

}
func main() {
	handlers.Common()
	handlers.Config()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/getCovidData", func(c echo.Context) error {

		// create a Client
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
		return c.JSON(http.StatusOK, stateData)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
