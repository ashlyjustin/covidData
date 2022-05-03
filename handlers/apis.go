package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		stateData, err = GetRedisData(ctx, keys, collection, redisCache)
		if err != nil {
			fmt.Println("Error in getting data from redis")
		}
	} else {
		stateData, err = GetRedisData(ctx, StateCodesList[:], collection, redisCache)
	}

	return stateData, nil
}
func GetStateData(c echo.Context) error {
	states := []map[int]string{{1: "AA"}, {2: "AB"}}
	return c.JSON(http.StatusOK, states)
}
func GetRedisData(ctx context.Context, keys []string, collection CollectionAPI, redisCache *redis.Client) ([]State, *echo.HTTPError) {
	var stateData []State
	for _, key := range keys {
		var singleState State
		val, err := redisCache.Get(ctx, string(key)).Bytes()
		if err != nil {
			cursor, err := collection.Find(ctx, bson.M{"StateCode": key})
			if err != nil {
				log.Errorf("Unable to find the state data : %v", err)
				return stateData,
					echo.NewHTTPError(http.StatusNotFound, ErrorMessage{Message: "unable to find the state"})
			}
			var tempStateData []State
			cursor.All(ctx, &tempStateData)

			fmt.Println(tempStateData[0])
			singleState = tempStateData[0]
			singleStateEncoding, err := json.Marshal(singleState)
			redisError := redisCache.Set(ctx, string(key), singleStateEncoding, 30*time.Minute).Err()
			if redisError != nil {
				fmt.Printf(redisError.Error())
				echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to set cache state"})
			}
			fmt.Println("from db ", singleState)
			stateData = append(stateData, singleState)

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

	return stateData, nil
}
