package main

import (
	"context"
	"covidApp/handlers"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// @title Echo Swagger Example API
// @version 1.0
// @description This is a sample server server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:1323
// @BasePath /
// @schemes http
var e = echo.New()
var mongoClient mongo.Client
var (
	c          *mongo.Client
	db         *mongo.Database
	stateCol   *mongo.Collection
	RedisCache *redis.Client
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
	createRedisCache()
	postCovidData()

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

func createRedisCache() {
	RedisCache = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong, err := RedisCache.Ping(context.Background()).Result()
	fmt.Println(pong, err)
	// if there has been an error setting the value
	// handle the error
	if err != nil {
		fmt.Println(err)
	}
}

type Links struct {
	Url  string
	Name string
}

func main() {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	h := &handlers.StateHandler{Col: stateCol, RedisClient: *RedisCache}
	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.GET("/getUserState", h.GetUserStateData)
	e.GET("/getStateData", h.GetStateData)
	e.GET("/*", HomeHandler)
	e.Logger.Fatal(e.Start(":1323"))
}

type TemplateRegistry struct {
	templates *template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func HomeHandler(c echo.Context) error {
	// Please note the the second parameter "home.html" is the template name and should
	// be equal to the value stated in the {{ define }} statement in "views/home.html"
	Link := []Links{
		{
			Name: "Get All State Data",
			Url:  "https://5977-2401-4900-1c09-acb8-4985-d5a-83bf-2e51.in.ngrok.io/getStateData",
		},
		{
			Name: "Get Your State Data",
			Url:  "https://5977-2401-4900-1c09-acb8-4985-d5a-83bf-2e51.in.ngrok.io/getUserState",
		},
	}
	return c.Render(http.StatusOK, "index.html", Link)

}
