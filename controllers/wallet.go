package controllers

import (
	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
)

func GetWalletBallance(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	// Find the user's wallet
	var wallet models.Wallet
	if err := database.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve wallet"})
	}

	return c.JSON(fiber.Map{
		"balance": wallet.Balance,
	})

}

func GetWalletHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	// Find the user's walletHistory
	var walletHistory []models.WalletHistory
	if err := database.DB.Where("user_id = ?", userID).Find(&walletHistory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve walletHistory"})
	}

	if len(walletHistory)==0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No walletHistory found"})
	}

	return c.JSON(fiber.Map{
		"walletHistory": walletHistory,
	})
}