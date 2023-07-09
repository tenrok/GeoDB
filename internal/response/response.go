package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseCode = uint

// Коды сообщений, отправляемых браузеру
const (
	CodeSuccess ResponseCode = 0 // Сообщение об удачном выполнении
	CodeError   ResponseCode = 1 // Возникла какая-то ошибка
)

type Response struct {
	Code   ResponseCode `json:"code"`
	Msg    string       `json:"msg,omitempty"`
	Result any          `json:"result,omitempty"`
}

// SendSuccess
func SendSuccess(c *gin.Context, msg string, result any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: msg, Result: result})
}

// SendError
func SendError(c *gin.Context, msg string, err error) {
	fmt.Println(err.Error())
	c.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeError, Msg: msg})
}
