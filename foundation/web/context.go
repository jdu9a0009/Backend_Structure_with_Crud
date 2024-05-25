package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type Context struct {
	*gin.Context
	Ctx         context.Context
	queryErrors []FieldError
	paramErrors []FieldError
}

func NewContext(context *gin.Context, ctx context.Context) *Context {
	return &Context{Context: context, Ctx: ctx}
}

func (c *Context) Respond(data interface{}, statusCode int) error {
	//if statusCode >= 400 {
	//	logger := NewLogger("logs")
	//
	//	if err := logger.WriteLog(c, data); err != nil {
	//		log.Println(err)
	//	}
	//}
	// ###############################
	//ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "foundation.web.respond")
	//defer span.End()

	// Set the status code for the request logger middleware.
	// If the context is missing this value, request the service
	// to be shutdown gracefully.
	v, ok := c.Ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("web value is missing from context!")
	}
	v.StatusCode = statusCode

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		c.AbortWithStatus(statusCode)
		return nil
	}

	c.JSON(statusCode, data)

	return nil
}

func (c *Context) RespondError(err error) error {

	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.

	if webErr, ok := Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		return c.Respond(er, webErr.Status)
	}

	// If not, the handler sent any arbitrary error value so use 500.
	er := ErrorResponse{
		Error: err.Error(),
	}
	return c.Respond(er, http.StatusInternalServerError)
}

func (c *Context) RespondMobileError(err error) error {

	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.
	if webErr, ok := Cause(err).(*Error); ok {
		errMsg := webErr.Err.Error()
		if webErr.Fields != nil {
			for _, f := range webErr.Fields {
				errMsg += fmt.Sprintf("(%s: %s)", f.Field, f.Error)
			}
		}
		er := MobileErrorResponse{
			Error: errMsg,
		}
		return c.Respond(er, webErr.Status)
	}

	// If not, the handler sent any arbitrary error value so use 500.
	er := ErrorResponse{
		Error: err.Error(),
	}
	return c.Respond(er, http.StatusInternalServerError)
}

func (c *Context) BindFunc(data interface{}, requiredFields ...string) error {
	err := c.ShouldBind(data)
	if err != nil {
		return NewRequestError(errors.Wrap(err, "parsing request data"), http.StatusBadRequest)
	}

	err = validateStruct(data, requiredFields...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) GetQueryFunc(dataType reflect.Kind, query string) interface{} {
	switch dataType {
	case reflect.Int:
		if value, ok := c.GetQuery(query); ok {
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				c.queryErrors = append(c.queryErrors, FieldError{
					Error: "query must be number!",
					Field: query,
				})
			}
			return &valueInt
		}
	case reflect.Float32:
		if value, ok := c.GetQuery(query); ok {
			valueFloat, err := strconv.ParseFloat(value, 32)
			if err != nil {
				c.queryErrors = append(c.queryErrors, FieldError{
					Error: "query must be float32!",
					Field: query,
				})
			}
			valueFloat32 := float32(valueFloat)
			return &valueFloat32
		}
	case reflect.Float64:
		if value, ok := c.GetQuery(query); ok {
			valueFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				c.queryErrors = append(c.queryErrors, FieldError{
					Error: "query must be float32!",
					Field: query,
				})
			}
			return &valueFloat
		}
	case reflect.String:
		if value, ok := c.GetQuery(query); ok {
			return &value
		}
	case reflect.Bool:
		if value, ok := c.GetQuery(query); ok {
			valueBool, err := strconv.ParseBool(value)
			if err != nil {
				c.queryErrors = append(c.queryErrors, FieldError{
					Error: "query must be boolean!",
					Field: query,
				})
			}

			return &valueBool
		}
	}

	return nil
}

func (c *Context) ValidQuery() *Error {
	if len(c.queryErrors) > 0 {
		return &Error{
			Err:    errors.New("some queries are not valid"),
			Status: http.StatusBadRequest,
			Fields: c.queryErrors,
		}
	}

	return nil
}

func (c *Context) GetParam(paramType reflect.Kind, param string) interface{} {
	value := c.Param(param)
	switch paramType {
	case reflect.Int:
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			c.paramErrors = append(c.paramErrors, FieldError{
				Error: "param must be number!",
				Field: param,
			})
		}

		return valueInt
	case reflect.String:
		if value == "" {
			c.paramErrors = append(c.paramErrors, FieldError{
				Error: "param not found",
				Field: param,
			})
		}

		return value
	}

	return nil
}

func (c *Context) ValidParam() *Error {
	if len(c.paramErrors) > 0 {
		return &Error{
			Err:    errors.New("some params are not valid"),
			Status: http.StatusBadRequest,
			Fields: c.paramErrors,
		}
	}

	return nil
}

func validateStruct(s interface{}, requiredFields ...string) error {
	structVal := reflect.Value{}
	if reflect.Indirect(reflect.ValueOf(s)).Kind() == reflect.Struct {
		structVal = reflect.Indirect(reflect.ValueOf(s))
	} else {
		return errors.New("input param should be a struct")
	}

	errFields := make([]FieldError, 0)

	structType := reflect.Indirect(reflect.ValueOf(s)).Type()
	fieldNum := structVal.NumField()

	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		fieldName := structType.Field(i).Name

		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			log.Print(isSet, fieldName, reflect.ValueOf(field))
			for _, f := range requiredFields {
				if f == fieldName {
					errFields = append(errFields, FieldError{
						Error: "field is required!",
						Field: fieldName,
					})
				}
			}
		}
	}

	if len(errFields) > 0 {
		return &Error{
			Err:    errors.New("required fields"),
			Fields: errFields,
			Status: http.StatusBadRequest,
		}
	}

	return nil
}
