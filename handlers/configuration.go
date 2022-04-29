package handlers

import "fmt"

func Config() {
	fmt.Println("from config")
}

type Configurations struct {
	AppName         string `env:"APP_NAME" env-default:"covidData"`
	AppEnv          string `env:"APP_ENV" env-default:"DEV"`
	ApiPort         string `env:"API_PORT" env-default:"8080"`
	RedisPort       string `env:"REDIS_PORT" env-default:"1521"`
	MongoPort       string `env:"Mongo_Port" env-default:"1523"`
	Host            string `env:"HOST" env-default:"localhost"`
	LogLevel        string `env:"LOG_LEVEL" env-default:"ERROR"`
	CovidDataUrl    string `yaml:"COVID_DATA_URL" env:"COVID_DATA_URL" env-default:"localhost/error"`
	UserLocationUrl string `env:"USER_LOCATION_URL" env-default:"localhost/error"`
}

var Cfg Configurations
