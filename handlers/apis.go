package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
)

var e = echo.New()

type StateHandler struct {
	Col         CollectionAPI
	RedisClient redis.Client
}

// HealthCheck godoc
// @Summary Show the status of covid data of all states.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /getStateData [get]

func (h *StateHandler) GetStateData(c echo.Context) error {
	stateData, httpError := findStateData(context.Background(), c.QueryParams(), h.Col, &h.RedisClient)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, stateData)
}
func findStateData(ctx context.Context, q url.Values, collection CollectionAPI, redisCache *redis.Client) ([]State, *echo.HTTPError) {
	var stateData []State
	var keys []string
	var err error
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	fmt.Println("filter is ", filter)
	value, StateCode := filter["StateCode"]
	if StateCode {
		key, e := json.Marshal(value)
		if e != nil {
			fmt.Println("State Code invalid type format")
		}
		keys = append(keys, string(key))
		stateData, err = getRedisData(ctx, keys, collection, redisCache)
		if err != nil {
			fmt.Println("Error in getting data from redis")
		}
	} else {
		stateData, err = getRedisData(ctx, StateCodesList[:], collection, redisCache)
	}

	return stateData, nil
}
func getRedisData(ctx context.Context, keys []string, collection CollectionAPI, redisCache *redis.Client) ([]State, *echo.HTTPError) {
	var stateData []State
	for _, key := range keys {
		var singleState State
		fmt.Println("key is", key)
		val, err := redisCache.Get(ctx, string(key)).Bytes()
		if err != nil {
			cursor, err := collection.Find(ctx, bson.M{"StateCode": key})
			if err != nil {
				log.Errorf("Unable to find the state data : %v", err)
				fmt.Println("not in redis state data")
				return stateData,
					echo.NewHTTPError(http.StatusNotFound, ErrorMessage{Message: "unable to find the state"})
			}
			var tempStateData []State
			cursor.All(ctx, &tempStateData)
			if len(tempStateData) > 0 {
				fmt.Println(tempStateData[0])
				singleState = tempStateData[0]
				singleStateEncoding, _ := json.Marshal(singleState)
				redisError := redisCache.Set(ctx, string(key), singleStateEncoding, 30*time.Minute).Err()
				if redisError != nil {
					fmt.Printf(redisError.Error())
					echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to set cache state"})
				}
				fmt.Println("from db ", singleState)
				stateData = append(stateData, singleState)
			}

		} else {
			fmt.Println(key, "from redis cache", string(val), " setting redis cache ###")
			parseError := json.Unmarshal(val, &singleState)
			if parseError != nil {
				log.Errorf("Unable to read the cursor : %v", parseError)
				return stateData,
					echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to parse retrieved state"})
			}

			stateData = append(stateData, singleState)

		}
	}
	fmt.Println("return value is ", stateData)
	return stateData, nil
}
func (h *StateHandler) GetUserStateData(c echo.Context) error {

	// ip:="223.177.38.252"
	// locationUrl:=Cfg.UserLocationUrl+ip+"?access_key="+Cfg.LocationApiKey§
	// stateData, httpError := findStateData(context.Background(), c.QueryParams(), h.Col, &h.RedisClient)
	// if httpError != nil {
	// 	return c.JSON(httpError.Code, httpError.Message)
	// }
	ip, err := getIp(c.Request().Header)
	if err != nil {
		fmt.Println(ip)
	}
	ip = "157.37.151.60"
	state, httpError := getUserState(ip, h.Col, &h.RedisClient)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}

	return c.JSON(http.StatusOK, state)
}
func getUserState(ip string, collection CollectionAPI, redisCache *redis.Client) (State, *echo.HTTPError) {
	queryUrl := Cfg.LocationApiKey + ip + "?access_key=" + Cfg.LocationApiKey
	queryUrl = "http://api.ipstack.com/223.177.38.252?access_key=59327922c4a73c01000c9a0391e89dfb&format=1"
	fmt.Println("queryUrl is", queryUrl)
	client := &http.Client{}
	var userState []State
	desc, err := client.Get(queryUrl)
	if err != nil {
		fmt.Print(err)
	}
	jsonData, err := ioutil.ReadAll(desc.Body)
	if err != nil {
		panic(err)
	}
	userIpData := make(map[string]json.RawMessage)
	e := json.Unmarshal(jsonData, &userIpData)
	if e != nil {
		fmt.Println("user ip data unavailable", e)
		return State{}, echo.NewHTTPError(http.StatusBadRequest, "State could not be found for the"+ip)
	}
	StateCode := string(userIpData["region_code"])
	StateCode = StateCode[1 : len(StateCode)-1]
	fmt.Println("Statecode is ", StateCode)
	key := []string{}
	key = append(key, StateCode)
	userState, iperror := getRedisData(context.Background(), key[:], collection, redisCache)
	if iperror != nil {
		if len(userState) < 1 {
			fmt.Println("data not found for DL")
			return State{}, echo.NewHTTPError(iperror.Code, "No data present for the state")
		}
	}
	fmt.Println(userState)

	return userState[0], nil

}
func getIp(req http.Header) (string, error) {
	ip := req.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := req.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("No valid ip found")
}
