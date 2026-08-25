package core_http_request

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

var requestValidator = validator.New()

type validatable interface {
	Validate() error
}

var (
	err error
)

func DecodeAndValidateRequest(c *gin.Context, dest any) error {
	if err := json.NewDecoder(c.Request.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	v, ok := dest.(validatable)
	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}

	if err != nil {
		return fmt.Errorf("request validation: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
