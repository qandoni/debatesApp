package core_http_request

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func GetIntPathValue(c *gin.Context, key string) (int, error) {
	pathValue := c.Param(key)
	if pathValue == "" {
		return 0, fmt.Errorf(
			"no key='%s' in path values: %w",
			key,
			core_errors.ErrInvalidArgument,
		)
	}
	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return 0, fmt.Errorf(
			"path value='%s' by key='%s' not a valid integer: %v: %w",
			pathValue,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return val, nil
}
