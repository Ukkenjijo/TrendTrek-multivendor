package controllers

import (
	"log"
	"strconv"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/Ukkenjijo/trendtrek/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func AdminLogin(c *fiber.Ctx) error {

	// Parse the request body into the LoginRequest struct
	req := new(models.EmailLoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validate the request input
	if err := utils.ValidateStruct(*req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch the user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Compare the hashed password with the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "You are not an authorized user"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}
	log.Println("tokenlogin", token)

	// Return the token and user info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
	})

}

func GetAllUsers(c *fiber.Ctx) error {

	type UserResponse struct {
		ID             uint   `json:"id"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		PhoneNumber    string `json:"phone_number"`
		ProfilePicture string `json:"picture"`
		Blocked        bool   `json:"blocked"`
	}

	var users []UserResponse

	if err := database.DB.Model(&models.User{}).
		Select("id, name, email, phone_number,profile_picture,blocked").
		Where("role = ?", models.RoleCustomer). // Assuming models.RoleCustomer contains the role 'customer'
		Find(&users).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Failed to retrieve data from the database, or the data doesn't exist",
		})
	}

	if len(users) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No customers found",
		})
	}
	// Return the list of customers
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"customers": users,
	})

}

func BlockUser(c *fiber.Ctx) error {
	// Get the user ID from the query parameters
	userIDStr := c.Query("userid")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}
	// Find the user by ID in the database
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	// Check if the user is already blocked
	if user.Blocked {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User is already blocked",
			"user_id": user.ID,
		})
	}

	// block the user
	user.Blocked = true
	// Save the updated user status to the database
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User has been blocked",
		"user_id": user.ID,
	})

}

// UnblockUser allows the admin to unblock a user by their user ID
func UnblockUser(c *fiber.Ctx) error {
	// Get the user ID from the query parameters
	userIDStr := c.Query("userid")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Convert the user ID to an integer
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Find the user by ID in the database
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if the user is already unblocked
	if !user.Blocked {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User is already unblocked",
			"user_id": user.ID,
		})
	}

	// Unblock the user
	user.Blocked = false

	// Save the updated user status to the database
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User has been unblocked",
		"user_id": user.ID,
	})
}

// ListBlockedUsers returns a list of users who are blocked
func ListBlockedUsers(c *fiber.Ctx) error {
	var blockedUsers []models.User

	// Query the database for users who are blocked
	if err := database.DB.Where("blocked = ?", true).Find(&blockedUsers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve blocked users",
		})
	}

	// If no blocked users are found, return a specific message
	if len(blockedUsers) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "No blocked users found",
			"users":   []models.User{},
		})
	}

	// Return the list of blocked users
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Blocked users retrieved successfully",
		"users":   blockedUsers,
	})
}

// AddCategory creates a new category
func AddCategory(c *fiber.Ctx) error {
	category := new(models.Category)

	// Parse the JSON request body
	if err := c.BodyParser(category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}
	if category.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category name is required",
		})
	}

	// Add category to the database
	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}

// EditCategory edits an existing category
func EditCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	var category models.Category

	// Find the category by ID
	if err := database.DB.First(&category, categoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Category not found",
		})
	}

	// Parse the JSON request body for updated fields
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}
	//check if the category is not repeated
	if err := database.DB.Where("name = ? AND id != ?", category.Name, category.ID).First(&models.Category{}).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category already exists",
		})
	}

	// Save the changes to the category
	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update category",
		})
	}

	return c.Status(fiber.StatusOK).JSON(category)
}

// DeleteCategory deletes a category by ID
func DeleteCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	var category models.Category

	// Find the category by ID
	if err := database.DB.First(&category, categoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Category not found",
		})
	}

	// Delete the category
	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete category",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}

// GetAllCategories returns all categories for the admin
func GetAllCategories(c *fiber.Ctx) error {
	var categories []models.Category
	var categoryResponses []models.CategoryResponse

	// Fetch all categories and preload their sub-categories
	if err := database.DB.Preload("ParentCategory").Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	// Map the categories to the custom response struct
	for _, category := range categories {
		if category.ParentCategoryID == nil {
			categoryResponse := mapCategoryToResponse(category, categories)
			categoryResponses = append(categoryResponses, categoryResponse)
		}
	}

	// If no categories are found, return an empty array
	if len(categories) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":    "No categories found",
			"categories": []models.Category{},
		})
	}

	// Return the list of categories
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Categories retrieved successfully",
		"categories": categoryResponses,
	})
}

// mapCategoryToResponse recursively maps a category to the response struct, including sub-categories
func mapCategoryToResponse(category models.Category, allCategories []models.Category) models.CategoryResponse {
	// Create a new CategoryResponse struct
	categoryResponse := models.CategoryResponse{
		ID:       category.ID,
		Name:     category.Name,
		IsActive: category.IsActive,
	}

	// Find and map sub-categories
	var subCategories []models.CategoryResponse
	for _, subCategory := range allCategories {
		if subCategory.ParentCategoryID != nil && *subCategory.ParentCategoryID == category.ID {
			subCategories = append(subCategories, mapCategoryToResponse(subCategory, allCategories))
		}
	}

	// Attach sub-categories, if any
	if len(subCategories) > 0 {
		categoryResponse.SubCategories = subCategories
	}

	return categoryResponse
}

func GetCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var categories models.Category

	// Fetch all categories
	if err := database.DB.First(&categories, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Category not Found",
		})
	}

	// Return the list of categories
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Category retrieved successfully",
		"category": categories,
	})
}

func UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")
    req:=new(models.StatusRequest)
	var order models.Order
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}
    if err:=database.DB.First(&order,orderID).Error;err!=nil{
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Order not found",
        })
    }
    order.Status=req.Status
    if err:=database.DB.Save(&order).Error;err!=nil{
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update order",
        })
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Order status updated successfully",
    })

}

