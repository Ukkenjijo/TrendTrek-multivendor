package controllers

import (
	"math"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
)

// SearchProducts handles the search query and returns a list of products
// that match the query parameters.
//
// Query parameters:
//
//	q (string): Search query (e.g., product name, description)
//	sort (string): Sorting criteria (e.g., popularity, price, ratings, featured, new_arrivals, alphabetical)
//	order (string): Sorting order (asc or desc)
//	category_id (int): Category ID (optional)
//	min_price (float): Minimum price (optional)
//	max_price (float): Maximum price (optional)
//	page (int): Page number (default is 1)
//	limit (int): Number of items per page (default is 20)
//
// Returns a JSON response with the following structure:
//
//	{
//	  "message": "Products retrieved successfully",
//	  "data": [...],
//	  "page": 1,
//	  "limit": 20,
//	  "total": 100,
//	  "total_pages": 5
//	}
func SearchProducts(c *fiber.Ctx) error {
	// Parse query parameters
	query := c.Query("q", "")               // Search query (e.g., product name, description)
	sort := c.Query("sort", "new_arrivals") // Default sorting is by new arrivals
	order := c.Query("order", "asc")        // Default order is ascending
	categoryID := c.QueryInt("category_id", 0)
	minPrice := c.QueryFloat("min_price", 0)
	maxPrice := c.QueryFloat("max_price", 0)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20) // Default to 20 items per page

    
    //validate page and limit
    if page < 1 || limit < 1 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Invalid page or limit",
        })
    }

	// Base query to search products
	db := database.DB.Model(&models.Product{}).Where("is_active = ?", true)

	// Apply search query (if provided)
	if query != "" {
		db = db.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	// Apply category filter (if provided)
	if categoryID != 0 {
		db = db.Where("category_id = ?", categoryID)
	}

	// Apply price range filter (if provided)
	if minPrice > 0 {
		db = db.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		db = db.Where("price <= ?", maxPrice)
	}

	// Sorting
	switch sort {
	case "popularity":
		db = db.Order("popularity_score " + order) // Assume there's a popularity score field
	case "price":
		db = db.Order("price " + order)
	case "ratings":
		db = db.Order("average_ratings " + order) // Assume there's an average ratings field
	case "featured":
		db = db.Order("is_featured " + order) // Assume there's a featured flag
	case "new_arrivals":
		db = db.Order("created_at " + order)
	case "alphabetical":
		db = db.Order("name " + order)
	default:
		db = db.Order("created_at " + order) // Default to sorting by new arrivals
	}

	// Pagination logic
	offset := (page - 1) * limit
	db = db.Offset(offset).Limit(limit)

	// Execute the query
	var products []models.Product
	var productResponses []models.ProductResponse
	if err := db.Preload("Category").Preload("Store").Preload("Images").Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search products"})
	}

	// Map products to the custom response struct
	for _, product := range products {
		productResponse := models.ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			StockQuantity: product.StockLeft,
			IsActive:      product.IsActive,
			Category: models.CategoryResponse{
				ID:   product.Category.ID,
				Name: product.Category.Name,
			},
			Store: models.StoreResponse{
				ID:   product.Store.ID,
				Name: product.Store.Name,
			},
			Images: make([]string, len(product.Images)),
		}

		// Map image URLs
		for i, image := range product.Images {
			productResponse.Images[i] = image.URL
		}

		// Add the product response to the list
		productResponses = append(productResponses, productResponse)
	}
	// Calculate total count of products
	var totalCount int64
	db.Model(&models.Product{}).Where("is_active = ?", true).Count(&totalCount)

	// Calculate total number of pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Send the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Products retrieved successfully",
		"data":    productResponses,
		"page":    page,
		"limit":   limit,
		"total_products":   totalCount,
        "total_pages": totalPages,
	})
}
