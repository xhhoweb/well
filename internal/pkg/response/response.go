package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/apperr"
)

// Response Standard API Response
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg,omitempty"`
}

// Success Success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: apperr.CodeSuccess,
		Data: data,
		Msg:  "success",
	})
}

// SuccessWithMsg Success with message
func SuccessWithMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: apperr.CodeSuccess,
		Data: data,
		Msg:  msg,
	})
}

// Fail Fail response with error
func Fail(c *gin.Context, err error) {
	if ae, ok := err.(*apperr.AppError); ok {
		c.JSON(http.StatusOK, Response{
			Code: ae.Code,
			Msg:  ae.Message,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: apperr.CodeInternalError,
		Msg:  err.Error(),
	})
}

// FailWithCode Fail with specific code
func FailWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// BadRequest Bad request response
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code: apperr.CodeBadRequest,
		Msg:  msg,
	})
}

// Unauthorized Unauthorized response
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: apperr.CodeUnauthorized,
		Msg:  msg,
	})
}

// NotFound Not found response
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code: apperr.CodeNotFound,
		Msg:  msg,
	})
}

// InternalError Internal server error response
func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: apperr.CodeInternalError,
		Msg:  msg,
	})
}
