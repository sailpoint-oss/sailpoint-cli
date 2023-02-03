package errors

import "fmt"

var ErrAccessTokenExpired = fmt.Errorf("accesstoken is expired")
