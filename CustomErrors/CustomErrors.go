package customerrors

import (
	"errors"
)

var ServerDownError = errors.New("Server is down")