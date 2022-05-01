package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

var e = echo.New()

type StateHandler struct {
	Col CollectionAPI
}

func (h *StateHandler) GetStateData(c echo.Context) error {
	stateData, httpError := findStateData(context.Background(), c.QueryParams(), h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, stateData)
}
func findStateData(ctx context.Context, q url.Values, collection CollectionAPI) ([]State, *echo.HTTPError) {
	var stateData []State
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	fmt.Println("filter is ", filter)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Errorf("Unable to find the state data : %v", err)
		return stateData,
			echo.NewHTTPError(http.StatusNotFound, ErrorMessage{Message: "unable to find the state"})
	}
	err = cursor.All(ctx, &stateData)
	if err != nil {
		log.Errorf("Unable to read the cursor : %v", err)
		return stateData,
			echo.NewHTTPError(http.StatusUnprocessableEntity, ErrorMessage{Message: "unable to parse retrieved state"})
	}
	return stateData, nil
}
func GetStateData(c echo.Context) error {
	states := []map[int]string{{1: "AA"}, {2: "AB"}}
	return c.JSON(http.StatusOK, states)
}
