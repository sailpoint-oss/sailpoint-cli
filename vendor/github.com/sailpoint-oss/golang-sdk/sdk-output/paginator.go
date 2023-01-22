package sailpoint

import (
	"net/http"
	"reflect"
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

		// check if this is the last set in the response. This could be enhanced by inspecting the header for the max results
		if (len(actualValue)) == 0 {
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
