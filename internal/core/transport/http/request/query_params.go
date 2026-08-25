package core_http_request

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func GetIntQueryParam(c *gin.Context, key string) (*int, error) {
	param := c.Request.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("param='%s' by key='%s' not a valid integer: %v, %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument)
	}

	return &val, nil
}

func GetLimitOffsetQueryParams(c *gin.Context) (*int, *int, error) {
	const (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)
	limit, err := GetIntQueryParam(c, limitQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'limit' query param: %w", err)
	}

	offset, err := GetIntQueryParam(c, offsetQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'offset' query param: %w", err)
	}
	return limit, offset, nil
}
