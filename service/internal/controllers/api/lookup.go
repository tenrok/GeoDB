package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"geodbsvc/internal/response"
)

// Lookup получает информацию об IP
func (c *Controller) Lookup() gin.HandlerFunc {
	db := c.srv.GetDB()
	validate := validator.New()

	return func(ctx *gin.Context) {
		ip := ctx.Param("ip")

		if err := validate.Var(ip, "required,ip"); err != nil {
			response.SendError(ctx, "Переданы неверные параметры", err)
			return
		}

		rec, err := db.Lookup(ip)
		if err != nil {
			response.SendError(ctx, "Возникла ошибка при получении информации", err)
			return
		}

		response.SendSuccess(ctx, "Информация получена успешно", rec)
	}
}
