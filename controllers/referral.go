package controllers

import (
	"fmt"
	"net/url"

	"github.com/Ukkenjijo/trendtrek/database"
	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/gofiber/fiber/v2"
	
)

func GenerateReferralLink(c *fiber.Ctx) error {
	user_id := c.Locals("user_id") // Assuming you have the user in context
	//Get the user from the database
	var user models.User
	if err := database.DB.Where("id = ?", user_id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user"})
	}
	baseURL := "http://locakhost:3000/api/v1/user/signup"
	referralLink := fmt.Sprintf("%s?referral_name=%s&referral_code=%s",
		baseURL, url.QueryEscape(user.Name), user.ReferralCode)

	return c.JSON(fiber.Map{
		"referral_link": referralLink,
	})
}

func ClaimReferral(referralName string, referralCode string, reffere uint) error {
	var user models.User
	if err := database.DB.Where("referral_code = ?", referralCode).First(&user).Error; err != nil {
		return err
	}

	// Get the wallet of the referrer
	referrerWallet := new(models.Wallet)
	err := database.DB.Where("user_id = ?", user.ID).First(&referrerWallet).Error
	if err != nil {
		return err
	}

	// Get the wallet of the referred user
	referredWallet := new(models.Wallet)
	err = database.DB.Where("user_id = ?", reffere).First(&referredWallet).Error
	if err != nil {
		return err
	}

	// Define the reward amount
	rewardAmount := 100.0

	// Update referrer's wallet balance and create history entry
	referrerWallet.Balance += rewardAmount
	history := models.WalletHistory{
		WalletID:  referrerWallet.ID,
		UserID:   user.ID,
		Amount:   rewardAmount,
		Operation: "credit",
		Balance:  referrerWallet.Balance,
		Reason:   "Referral reward",
	}
	database.DB.Model(&referrerWallet).Updates(referrerWallet)
	database.DB.Create(&history)

	// Update referred user's wallet balance and create history entry
	referredWallet.Balance += rewardAmount
	history = models.WalletHistory{
		WalletID:  referredWallet.ID,
		UserID:   reffere,
		Amount:   rewardAmount,
		Operation: "credit",
		Balance:  referredWallet.Balance,
		Reason:   "Referral bonus",
	}
	database.DB.Model(&referredWallet).Updates(referredWallet)
	database.DB.Create(&history)

	return nil
}