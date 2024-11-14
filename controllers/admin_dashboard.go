package controllers

import (
	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/gofiber/fiber/v2"
)

type TopProduct struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type TopCategory struct {
	CategoryID   uint    `json:"category_id"`
	Name         string  `json:"name"`
	TotalRevenue float64 `json:"total_revenue"`
}

type TopSeller struct {
	SellerID     uint    `json:"seller_id"`
	Name         string  `json:"name"`
	TotalRevenue float64 `json:"total_revenue"`
}

// GetTopProducts returns the top 10 products based on quantity sold
func GetTopProducts(c *fiber.Ctx) error {
	var topProducts []TopProduct

	err := database.DB.Table("order_items").
		Select("products.id as product_id, products.name, SUM(order_items.quantity) as total_sold, SUM(order_items.total_price) as total_revenue").
		Joins("JOIN products ON order_items.product_id = products.id").
		Group("products.id").
		Order("total_sold DESC").
		Limit(10).
		Scan(&topProducts).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch top products"})
	}

	return c.JSON(fiber.Map{"top_products": topProducts})
}
// GetTopCategories returns the top 10 categories based on total revenue
func GetTopCategories(c *fiber.Ctx) error {
	var topCategories []TopCategory

	err := database.DB.Table("order_items").
		Select("categories.id as category_id, categories.name, SUM(order_items.total_price) as total_revenue").
		Joins("JOIN products ON order_items.product_id = products.id").
		Joins("JOIN categories ON products.category_id = categories.id").
		Group("categories.id").
		Order("total_revenue DESC").
		Limit(10).
		Scan(&topCategories).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch top categories"})
	}

	return c.JSON(fiber.Map{"top_categories": topCategories})
}

// GetTopSellers returns the top 10 sellers based on total revenue
func GetTopSellers(c *fiber.Ctx) error {
	var topSellers []TopSeller
	if err:=database.DB.Table("order_items").
		Select("stores.id as store_id, stores.name, SUM(order_items.total_price) as total_revenue").
		Joins("JOIN products ON order_items.product_id = products.id").
		Joins("JOIN stores ON products.store_id = stores.id").
		Group("stores.id").
		Order("total_revenue DESC").
		Limit(10).
		Scan(&topSellers).Error; err!=nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch top sellers"+err.Error()})
	}
	return c.JSON(fiber.Map{"top_sellers": topSellers}) 
}