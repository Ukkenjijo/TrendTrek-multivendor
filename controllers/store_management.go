package controllers

import (
	"fmt"
	"log"

	"strconv"
	"time"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AddProduct handles the creation of a new product with an image upload
func AddProduct(c *fiber.Ctx) error {
	// Parse multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse multipart form",
		})
	}

	// Extract product fields from the form
	name := form.Value["name"][0]
	description := form.Value["description"][0]
	price, _ := strconv.ParseFloat(form.Value["price"][0], 64)
	stockQuantity, _ := strconv.Atoi(form.Value["stock_quantity"][0])
	storeID, _ := strconv.Atoi(form.Value["store_id"][0])
	categoryID, _ := strconv.Atoi(form.Value["category_id"][0])

	// Create a new Product struct
	product := models.Product{
		Name:          name,
		Description:   description,
		Price:         price,
		StockQuantity: stockQuantity,
		StoreID:       uint(storeID),
		CategoryID:    uint(categoryID),
		IsActive:      true,
		StockLeft:     stockQuantity,
	}

	// Handle multiple image uploads
	files := form.File["images"] // "images" is the name attribute in the form

	for _, file := range files {
		// Generate a unique file name for each image
		timestamp := time.Now().Unix()
		fileName := strconv.Itoa(int(product.StoreID)) + "_" + strconv.Itoa(int(timestamp)) + "_" + file.Filename

		// Save the image to the "uploads" directory
		filePath := "./uploads/product_images/" + fileName
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save image",
			})
		}

		// Store the image URL (assuming you're serving the images statically)
		imageURL := fmt.Sprintf("http://localhost:3000/uploads/product_images/%s", fileName)

		// Append each image URL to the product's Images array
		image := models.Image{
			URL: imageURL,
		}
		product.Images = append(product.Images, image)
	}

	// Validate the request input
	if err := utils.ValidateStruct(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Save the product to the database
	if err := database.DB.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// AddProduct handles the creation of a new product with an image upload
func EditProduct(c *fiber.Ctx) error {

	productID := c.Params("id")
	product := new(models.Product)

	// Find the product by ID
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Parse multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse multipart form",
		})
	}

	// Update product fields
	product.Name = form.Value["name"][0]
	product.Description = form.Value["description"][0]
	product.Price, _ = strconv.ParseFloat(form.Value["price"][0], 64)
	product.StockQuantity, _ = strconv.Atoi(form.Value["stock_quantity"][0])
	StockLeft, _ := strconv.Atoi(form.Value["stock_quantity"][0])
	product.StockLeft += StockLeft
	categoryID, _ := strconv.Atoi(form.Value["category_id"][0])
	product.CategoryID = uint(categoryID)

	// Handle multiple image uploads
	files := form.File["images"] // "images" is the name attribute in the form

	for _, file := range files {
		// Generate a unique file name for each image
		timestamp := time.Now().Unix()
		fileName := strconv.Itoa(int(product.StoreID)) + "_" + strconv.Itoa(int(timestamp)) + "_" + file.Filename

		// Save the image to the "uploads" directory
		filePath := "./uploads/product_images/" + fileName
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save image",
			})
		}

		// Store the image URL (assuming you're serving the images statically)
		imageURL := fmt.Sprintf("http://localhost:3000/uploads/product_images/%s", fileName)

		// Append each image URL to the product's Images array
		image := models.Image{
			URL: imageURL,
		}
		product.Images = append(product.Images, image)
	}

	// Validate the request input
	if err := utils.ValidateStruct(product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	/// Save the updated product to the database
	if err := database.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// DeleteProduct handles soft deleting a product
func DeleteProduct(c *fiber.Ctx) error {
	productID := c.Params("id")
	var product models.Product

	// Find the product by ID
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Perform the soft delete
	if err := database.DB.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete product",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}

// GetAllProducts fetches all products and maps them to the custom response struct
func GetAllProducts(c *fiber.Ctx) error {
	var products []models.Product
	var productResponses []models.ProductResponse

	// Query to fetch all products with related Category, Store, and Images
	if err := database.DB.Preload("Category").Preload("Store").Preload("Images").Preload("Offer").Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Map products to the custom response struct
	for _, product := range products {
		productResponse := models.ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			StockQuantity: product.StockQuantity,
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
		//Calculate discount price for each product
		if product.Offer != nil && product.Offer.DiscountPercentage > 0 {
			discountPercentage := product.Offer.DiscountPercentage
			discountedPrice := product.Price * (1 - discountPercentage/100)

			productResponse.DiscountPercentage = &discountPercentage
			productResponse.DiscountedPrice = &discountedPrice
		}

		// Add the product response to the list
		productResponses = append(productResponses, productResponse)
	}

	// Return the list of products with the custom response struct
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Products retrieved successfully",
		"products": productResponses,
	})
}

func GetProductbyId(c *fiber.Ctx) error {
	id := c.Params("id") // retrieve the id parameter
	var product models.Product
	var productResponse models.ProductResponse

	// Query to fetch all products with related Category, Store, and Images
	if err := database.DB.Preload("Category").Preload("Store").Preload("Images").First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch product",
		})
	}

	productResponse = models.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
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

	// Return the list of products with the custom response struct
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Products retrieved successfully",
		"product": productResponse,
	})

}

func GetStoreIDByUserID(userID uint) (uint, error) {
	var store models.Store

	err := database.DB.Model(&store).Where("user_id = ?", userID).First(&store).Error
	if err != nil {
		return 0, err
	}

	return store.ID, nil
}

func GetProducts(c *fiber.Ctx) error {
	// Assume you have a way to get the logged-in seller's ID from the context or session
	userID := c.Locals("user_id") // Replace with your actual implementation
	sellerID, _ := GetStoreIDByUserID(uint(userID.(float64)))
	log.Println(sellerID)

	var products []models.Product
	var productResponses []models.ProductResponse

	// Query to fetch products of the logged-in seller with related Category, Store, and Images
	if err := database.DB.Where("store_id = ?", sellerID).
		Preload("Category").
		Preload("Store").
		Preload("Images").
		Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Map products to the custom response struct
	for _, product := range products {
		productResponse := models.ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			StockQuantity: product.StockQuantity,
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

	// Return the list of products with the custom response struct
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Products retrieved successfully",
		"products": productResponses,
	})
}

func GetProductsByCategory(c *fiber.Ctx) error {
	categoryID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category_id",
		})
	}
	log.Println(categoryID)

	var products []models.Product
	var productResponses []models.ProductResponse

	// Query to fetch products of the specified category with related Category, Store, and Images
	if err := database.DB.Preload("Category").Preload("Store").Preload("Images").
		Where("category_id = ?", uint(categoryID)).Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Map products to the custom response struct
	for _, product := range products {
		productResponse := models.ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			StockQuantity: product.StockQuantity,
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

	// Return the list of products with the custom response struct
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Products retrieved successfully",
		"products": productResponses,
	})
}

// update the product stock
func UpdateProductStock(c *fiber.Ctx) error {
	// Get the product ID from the URL
	id := c.Params("id")
	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}
	log.Println(product.StockQuantity)

	// Get the new stock quantity from the request body
	var request models.UpdateProductStockRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	stockQuantity := request.StockQuantity
	// Update the product stock quantity
	product.StockQuantity += stockQuantity
	if err := database.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product stock",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product stock updated successfully",
	})
}

func ReducestockandDeleteCart(tx *gorm.DB, cart *models.Cart) error {
	var cartItems []models.CartItem
	if err := tx.Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		return err
	}
	for _, cartItem := range cartItems {
		var product models.Product
		if err := tx.Where("id = ?", cartItem.ProductID).First(&product).Error; err != nil {
			return err
		}
		product.StockQuantity -= cartItem.Quantity
		if err := tx.Save(&product).Error; err != nil {
			return err
		}
	}
	if err := tx.Delete(cart).Error; err != nil {
		return err
	}
	return nil
}
