package main

import (
	"covidApp/handlers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var e = echo.New()

func init() {
	fmt.Println("inside init")
	err := cleanenv.ReadConfig("config.yml", &handlers.Cfg)
	if err != nil {
		e.Logger.Fatal("Unable to load configuration")
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
			var total handlers.TotalData
			tree := make(map[string]json.RawMessage)
			fmt.Println(string(value))
			json.Unmarshal(value, &tree)
			json.Unmarshal(tree["total"], &total)
			state := handlers.State{StateCode: key, Total: total}
			stateData = append(stateData, state)
		}
		fmt.Println(stateData)
		return c.JSON(http.StatusOK, stateData)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
