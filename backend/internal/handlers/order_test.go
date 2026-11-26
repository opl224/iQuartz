package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder_InsufficientStock_NothingWritten(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := map[string]interface{}{
		"customer": "Test User",
		"items": []map[string]interface{}{
			{"sku": "TEST-ZERO-STOCK", "qty": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateOrder(testDB)
	handler(c)

	assert.Equal(t, 409, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_STOCK")
}

func TestCreateOrder_HappyPath_WritesMovementAndUpdatesStock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := map[string]interface{}{
		"customer": "Test User",
		"items": []map[string]interface{}{
			{"sku": "LAPTOP-001", "qty": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateOrder(testDB)
	handler(c)

	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["id"])
	assert.NotNil(t, response["total"])
}