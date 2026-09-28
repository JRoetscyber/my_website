package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type OrderHandler struct {
	DB           *gorm.DB
	Cfg          *config.Config
	AdminHandler *AdminHandler
}

func NewOrderHandler(db *gorm.DB, cfg *config.Config, adminHandler *AdminHandler) *OrderHandler {
	return &OrderHandler{
		DB:           db,
		Cfg:          cfg,
		AdminHandler: adminHandler,
	}
}

// OrderTracker renders the customer-facing order tracking page with live updates & notifications
func (h *OrderHandler) OrderTracker(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return c.Status(http.StatusNotFound).SendString("Invalid order tracking link")
	}

	var order models.Order
	if err := h.DB.Where("tracking_code = ?", code).First(&order).Error; err != nil {
		return c.Status(http.StatusNotFound).Render("order_tracker", fiber.Map{
			"error":   "Order not found. Please check your tracking link.",
			"request": c,
		})
	}

	return c.Render("order_tracker", fiber.Map{
		"order":   order,
		"request": c,
	})
}

// OrderStatusAPI returns real-time JSON status for the customer's web notification listener
func (h *OrderHandler) OrderStatusAPI(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	var order models.Order
	if err := h.DB.Where("tracking_code = ?", code).First(&order).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status": "error",
			"error":  "Order not found",
		})
	}

	return c.JSON(fiber.Map{
		"status":          "success",
		"id":              order.ID,
		"sequence_number": order.SequenceNumber,
		"tracking_code":   order.TrackingCode,
		"customer_name":   order.CustomerName,
		"order_status":    order.Status,
		"items":           order.Items,
		"total_amount":    order.TotalAmount,
		"notes":           order.Notes,
		"notify_count":    order.NotifyCount,
		"ready_at":        order.ReadyAt,
		"created_at":      order.CreatedAt,
	})
}

// OrderDisplay renders the full-screen counter / TV Call-Out screen
func (h *OrderHandler) OrderDisplay(c *fiber.Ctx) error {
	return c.Render("order_display", fiber.Map{
		"request": c,
	})
}

// OrderDisplayDataAPI returns active preparing and ready orders for the counter display
func (h *OrderHandler) OrderDisplayDataAPI(c *fiber.Ctx) error {
	today := time.Now().Add(-14 * time.Hour)

	var preparingOrders []models.Order
	h.DB.Where("created_at >= ? AND status IN ('received', 'in_progress')", today).
		Order("sequence_number asc").Find(&preparingOrders)

	var readyOrders []models.Order
	h.DB.Where("created_at >= ? AND status = 'ready'", today).
		Order("ready_at desc, sequence_number asc").Limit(16).Find(&readyOrders)

	return c.JSON(fiber.Map{
		"preparing": preparingOrders,
		"ready":     readyOrders,
	})
}

// AdminOrdersPage renders the orders dashboard in the admin portal
func (h *OrderHandler) AdminOrdersPage(c *fiber.Ctx) error {
	var orders []models.Order
	h.DB.Order("id desc").Limit(100).Find(&orders)

	var preparingCount int64
	var readyCount int64
	var completedCount int64
	today := time.Now().Truncate(24 * time.Hour)

	h.DB.Model(&models.Order{}).Where("status IN ('received', 'in_progress')").Count(&preparingCount)
	h.DB.Model(&models.Order{}).Where("status = 'ready'").Count(&readyCount)
	h.DB.Model(&models.Order{}).Where("status = 'completed' AND created_at >= ?", today).Count(&completedCount)

	nextSeq := services.NextCalloutSequence(h.DB)

	ctx := h.AdminHandler.buildAdminContext(c, "orders", fiber.Map{
		"orders":          orders,
		"preparing_count": preparingCount,
		"ready_count":     readyCount,
		"completed_count": completedCount,
		"next_sequence":   nextSeq,
		"today_date":      time.Now().Format("02 Jan 2006"),
	})

	return c.Render("admin/orders", ctx)
}

// AdminCreateOrder creates a new order, calculates next sequence ticket, and generates tracking code
func (h *OrderHandler) AdminCreateOrder(c *fiber.Ctx) error {
	customerName := strings.TrimSpace(c.FormValue("customer_name"))
	if customerName == "" {
		customerName = "Guest Customer"
	}
	customerPhone := strings.TrimSpace(c.FormValue("customer_phone"))
	customerEmail := strings.TrimSpace(c.FormValue("customer_email"))
	items := strings.TrimSpace(c.FormValue("items"))
	notes := strings.TrimSpace(c.FormValue("notes"))
	totalAmount, _ := strconv.ParseFloat(c.FormValue("total_amount"), 64)

	sequenceNum := services.NextCalloutSequence(h.DB)
	trackingCode := services.GenerateTrackingCode()

	order := models.Order{
		SequenceNumber: sequenceNum,
		TrackingCode:   trackingCode,
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		CustomerEmail:  customerEmail,
		Items:          items,
		TotalAmount:    totalAmount,
		Status:         "received",
		Notes:          notes,
	}

	if err := h.DB.Create(&order).Error; err != nil {
		log.Printf("[ORDERS] Failed to create order: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Failed to create order")
	}

	log.Printf("[ORDERS] Created Order #%d for '%s' (Tracking: %s)", order.SequenceNumber, order.CustomerName, order.TrackingCode)

	if strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{
			"status": "success",
			"order":  order,
		})
	}

	return c.Redirect("/admin/orders")
}

// AdminUpdateOrderStatus updates order status and triggers ready notifications
func (h *OrderHandler) AdminUpdateOrderStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid order ID")
	}

	newStatus := strings.ToLower(strings.TrimSpace(c.FormValue("status")))
	if newStatus == "" {
		var body map[string]string
		if err := c.BodyParser(&body); err == nil {
			newStatus = strings.ToLower(strings.TrimSpace(body["status"]))
		}
	}

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		return c.Status(http.StatusNotFound).SendString("Order not found")
	}

	order.Status = newStatus
	now := time.Now()

	if newStatus == "ready" {
		order.ReadyAt = &now
		order.NotifyCount++
		log.Printf("[ORDERS] Order #%d marked READY! Triggered customer notification #%d", order.SequenceNumber, order.NotifyCount)
	} else if newStatus == "completed" {
		order.CompletedAt = &now
	}

	h.DB.Save(&order)

	if strings.Contains(c.Get("Accept"), "application/json") || c.Is("json") {
		return c.JSON(fiber.Map{
			"status": "success",
			"order":  order,
		})
	}

	return c.Redirect("/admin/orders")
}

// AdminNotifyOrder manually re-triggers the callout alert for a ready order
func (h *OrderHandler) AdminNotifyOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid order ID")
	}

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		return c.Status(http.StatusNotFound).SendString("Order not found")
	}

	order.Status = "ready"
	now := time.Now()
	if order.ReadyAt == nil {
		order.ReadyAt = &now
	}
	order.NotifyCount++
	h.DB.Save(&order)

	log.Printf("[ORDERS] Re-notified callout for Order #%d (Notification count: %d)", order.SequenceNumber, order.NotifyCount)

	if strings.Contains(c.Get("Accept"), "application/json") || c.Is("json") {
		return c.JSON(fiber.Map{
			"status":       "success",
			"notify_count": order.NotifyCount,
			"order":        order,
		})
	}

	return c.Redirect("/admin/orders")
}

// AdminDeleteOrder deletes an order
func (h *OrderHandler) AdminDeleteOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid order ID")
	}

	h.DB.Delete(&models.Order{}, id)
	log.Printf("[ORDERS] Deleted order ID %d", id)

	if strings.Contains(c.Get("Accept"), "application/json") || c.Is("json") {
		return c.JSON(fiber.Map{"status": "success"})
	}

	return c.Redirect("/admin/orders")
}
