package sailpoint

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"

	v3 "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
)

func PaginateWithDefaults[T any](f interface{}) ([]T, *http.Response, error) {
	return Paginate[T](f, 0, 250, 10000)
}

func Paginate[T any](f interface{}, initialOffset int32, increment int32, limit int32) ([]T, *http.Response, error) {
	var offset int32 = initialOffset
	var returnObject []T
	var latestResponse *http.Response
	for offset < limit {
		// first invoke the Offset command to set the new offset
		offsetResp := Invoke(f, "Offset", offset)
		// invoke the Execute function to get the response
		resp := Invoke(offsetResp[0].Interface(), "Execute")

		// convert the expected return values to their respective types
		actualValue := resp[0].Interface().([]T)
		latestResponse = resp[1].Interface().(*http.Response)
		err := resp[2].Interface()

		if err != nil {
			return returnObject, latestResponse, err.(error)
		}

		// append the results to the main return object
		returnObject = append(returnObject, actualValue...)

		// check if this is the last set in the response. This could be enhanced by inspecting the header for the max results
		if int32(len(actualValue)) < increment {
			break
		}

		offset += increment
	}
	return returnObject, latestResponse, nil
}

func PaginateSearchApi(ctx context.Context, apiClient *APIClient, search v3.Search, initialOffset int32, increment int32, limit int32) ([]map[string]interface{}, *http.Response, error) {
	var offset int32 = initialOffset
	var returnObject []map[string]interface{}
	var latestResponse *http.Response

	if len(search.Sort) != 1 {
		return nil, nil, errors.New("search must include exactly one sort parameter to paginate properly")
	}

	for offset < limit {
		if len(returnObject) > 0 {
			search.SearchAfter = []string{returnObject[len(returnObject)-1][strings.Trim(search.Sort[0], "-")].(string)}
		}
		// convert the expected return values to their respective types
		actualValue, latestResponse, err := apiClient.V3.SearchApi.SearchPost(ctx).Limit(increment).Search(search).Execute()

		if err != nil {
			return returnObject, latestResponse, err.(error)
		}

		// check if this is the last set in the response. This could be enhanced by inspecting the header for the max results
		if int32(len(actualValue)) < increment {
			break
		}

		// append the results to the main return object
		returnObject = append(returnObject, actualValue...)
		offset += increment
	}
	return returnObject, latestResponse, nil
}

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}
