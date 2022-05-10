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
func findStateData(ctx context.Context, q url.Values, collection CollectionAPI, redisCache *redis.Client) ([]SingleState, *echo.HTTPError) {
	var stateData []SingleState
	var keys []string
	var err error
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	value, StateCode := filter["StateCode"]
	if StateCode {
		key, e := json.Marshal(value)
		if e != nil {
			fmt.Println("State Code invalid type format")
		}
		keys = append(keys, string(key))
		stateData, err = getRedisData(ctx, keys, collection, redisCache)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		stateData, err = getRedisData(ctx, StateCodesList[:], collection, redisCache)
	}

	return stateData, nil
}
func getRedisData(ctx context.Context, keys []string, collection CollectionAPI, redisCache *redis.Client) ([]SingleState, *echo.HTTPError) {
	var stateData []SingleState
	for _, key := range keys {
		var singleState State
		val, err := redisCache.Get(ctx, string(key)).Bytes()
		if err != nil {
			cursor, err := collection.FindOne(ctx, bson.M{"StateCode": key}).DecodeBytes()
			if err != nil {
				log.Errorf("Unable to find the state data : %v", err)
				return stateData,
					echo.NewHTTPError(http.StatusNotFound, ErrorMessage{Message: "unable to find the state"})
			}
			// var tempStateData []State
			parseError := bson.Unmarshal(cursor, &singleState)
			if parseError != nil {
				log.Errorf("Unable to read the cursor : %v", parseError)
				return stateData,
					echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to parse retrieved state"})
			} else {
				singleStateEncoding, _ := json.Marshal(singleState)
				redisError := redisCache.Set(ctx, string(key), singleStateEncoding, 30*time.Minute).Err()
				if redisError != nil {
					fmt.Println(redisError.Error())
					echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to update cache state"})
				}
				stateData = append(stateData, SingleState{State: singleState, StateName: StateNameMap[string(key)]})
			}
			// if len(tempStateData) > 0 {
			// 	singleState = tempStateData[0]
			// 	singleStateEncoding, _ := json.Marshal(singleState)
			// 	redisError := redisCache.Set(ctx, string(key), singleStateEncoding, 30*time.Minute).Err()
			// 	if redisError != nil {
			// 		fmt.Println(redisError.Error())
			// 		echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to set cache state"})
			// 	}

			// 	stateData = append(stateData, SingleState{State: singleState, StateName: StateNameMap[string(key)]})
			// }

		} else {
			fmt.Println("Got data from cache redis")
			parseError := json.Unmarshal(val, &singleState)
			if parseError != nil {
				log.Errorf("Unable to read the cursor : %v", parseError)
				return stateData,
					echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to parse retrieved state"})
			}
			stateData = append(stateData, SingleState{State: singleState, StateName: StateNameMap[string(key)]})

		}
	}
	return stateData, nil
}

func (h *StateHandler) GetUserStateData(c echo.Context) error {
	ip, err := getIp(c.Request().Header)
	if err != nil {
		fmt.Println(err.Error())
		return c.HTML(http.StatusForbidden, "Request from source invalid")
	}
	userState, httpError := getUserState(ip, h.Col, &h.RedisClient)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}

	return c.JSON(http.StatusOK, userState)
}
func getUserState(ip string, collection CollectionAPI, redisCache *redis.Client) (UserState, *echo.HTTPError) {
	queryUrl := Cfg.UserLocationUrl + ip
	client := &http.Client{}
	var userState []SingleState
	desc, err := client.Get(queryUrl)
	if err != nil {
		fmt.Print(err)
	}
	jsonData, err := ioutil.ReadAll(desc.Body)
	if err != nil {
		panic(err)
	}
	defer desc.Body.Close()
	// userIpData := make(map[string]json.RawMessage)
	var geoLocationData GeoLocation
	e := json.Unmarshal(jsonData, &geoLocationData)
	if e != nil {
		fmt.Println("user ip data unavailable", e)
		return UserState{Location: GeoLocation{}, State: SingleState{}}, echo.NewHTTPError(http.StatusBadRequest, "State could not be found for the"+ip)
	}
	StateCode := geoLocationData.Region
	// StateCode = StateCode[1 : len(StateCode)-1]
	key := []string{}
	key = append(key, StateCode)
	fmt.Println("key is", StateCode)
	userState, iperror := getRedisData(context.Background(), key[:], collection, redisCache)
	if iperror != nil {
		if len(userState) < 1 {
			fmt.Println("data not found for ", StateCode)
			return UserState{Location: geoLocationData, State: SingleState{State: State{}, StateName: StateNameMap[StateCode]}}, echo.NewHTTPError(iperror.Code, "No data present for the state")
		}
	}
	fmt.Println(userState)

	return UserState{Location: geoLocationData, State: userState[0]}, nil

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

	return "", fmt.Errorf(" no valid ip found")
}
