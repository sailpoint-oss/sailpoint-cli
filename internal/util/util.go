package util

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return (string(b))
}
