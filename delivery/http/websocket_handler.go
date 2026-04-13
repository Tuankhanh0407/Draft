// Import appropriate package.
package http

// Import necessary libraries.
import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

// WebSocketHandler manages real-time WebSocket connections for the application.
type WebSocketHandler struct {
	RedisClient *redis.Client
}

// NewWebSocketHandler initializes endpoints for WebSocket streaming.
func NewWebSocketHandler(app *fiber.App, rdb *redis.Client) {
	handler := &WebSocketHandler{
		RedisClient: rdb,
	}
	// Middleware to check if the request is a valid WebSocket upgrade.
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	// Mount the WebSocket route.
	app.Get("/ws/v1/exams/:id/live", websocket.New(handler.LiveDashboard))
}

// LiveDashboard streams real-time submission events to connected clients (teachers/admins).
// @Summary Connect to live dashboard WebSocket
// @Description Upgrade HTTP connection to WebSocket to receive real-time submission events (WS client required, not testable via Swagger UI directly).
// @Tags Dashboards
// @Param id path int true "Exam ID"
// @Success 101 {string} string "Switching protocols to WebSocket"
// @Failure 426 {string} string "Update required"
// @Router /ws/v1/exams/{id}/live [get]
func (h *WebSocketHandler) LiveDashboard(c *websocket.Conn) {
	examIDStr := c.Params("id")
	channelName := fmt.Sprintf("exam_live_dashboard_%s", examIDStr)
	// 1. Subscribe to the specific Redis Pub/Sub channel for this exam.
	ctx := context.Background()
	pubsub := h.RedisClient.Subscribe(ctx, channelName)
	defer pubsub.Close() // Ensure subcription is closed when the WebSocket drops.
	// 2. Notify the client that connection is successful.
	welcomeMsg := fiber.Map{
		"status":	"connected",
		"message":	"Listening to live events for exam " + examIDStr,
	}
	if err := c.WriteJSON(welcomeMsg); err != nil {
		log.Println("WebSocket write error:", err)
		return
	}
	// 3. Infinite loop to listen for Redis events and push them to the WebSocket client.
	ch := pubsub.Channel()
	for msg := range ch {
		// msg.Payload contains the JSON string generated in SubmissionUseCase.EvaluateAndSave
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Println("Client disconnected or error writing to WebSocket:", err)
			break // Break the loop and close connection if the client disconnects.
		}
	}
}