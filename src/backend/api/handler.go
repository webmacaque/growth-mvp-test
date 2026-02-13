package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"growth-mvp/backend/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *domain.Service
}

func NewHandler(service *domain.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/shops/:shopId/telegram/connect", h.connectTelegram)
	router.POST("/shops/:shopId/orders", h.createOrder)
	router.GET("/shops/:shopId/telegram/status", h.telegramStatus)
}

func (h *Handler) connectTelegram(c *gin.Context) {
	shopID, ok := parseShopID(c)
	if !ok {
		return
	}

	var input domain.ConnectTelegramInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(input.BotToken) == "" || strings.TrimSpace(input.ChatID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "botToken and chatId must be non-empty"})
		return
	}

	input.BotToken = strings.TrimSpace(input.BotToken)
	input.ChatID = strings.TrimSpace(input.ChatID)

	out, err := h.service.ConnectTelegram(c.Request.Context(), shopID, input)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handler) createOrder(c *gin.Context) {
	shopID, ok := parseShopID(c)
	if !ok {
		return
	}

	var input domain.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Number = strings.TrimSpace(input.Number)
	input.CustomerName = strings.TrimSpace(input.CustomerName)

	out, err := h.service.CreateOrder(c.Request.Context(), shopID, input)
	if err != nil {
		if errors.Is(err, domain.ErrShopNotIntegrated) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *Handler) telegramStatus(c *gin.Context) {
	shopID, ok := parseShopID(c)
	if !ok {
		return
	}
	out, err := h.service.GetTelegramStatus(c.Request.Context(), shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func parseShopID(c *gin.Context) (int64, bool) {
	raw := c.Param("shopId")
	shopID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || shopID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shopId"})
		return 0, false
	}
	return shopID, true
}
