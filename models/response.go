package models

// ProductResponse represents the structure of the product data sent to the client
type ProductResponse struct {
	ID                 uint             `json:"id"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	Price              float64          `json:"price"`
	StockQuantity      int              `json:"stock_quantity"`
	DiscountedPrice    *float64         `json:"discounted_price,omitempty"`    // Discounted price if offer exists
	DiscountPercentage *float64         `json:"discount_percentage,omitempty"` // Discount percentage if offer exists
	IsActive           bool             `json:"is_active"`
	Category           CategoryResponse `json:"category"`
	Store              StoreResponse    `json:"store"`
	Images             []string         `json:"images"` // List of image URLs
}
type CategoryResponse struct {
	ID            uint               `json:"id"`
	Name          string             `json:"name"`
	IsActive      bool               `json:"is_active"`
	SubCategories []CategoryResponse `json:"sub_categories,omitempty"` // Sub-categories, if any
}

// StoreResponse represents the structure of the store in the product response
type StoreResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type UserProfileResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone"`
	ProfilePicture string `json:"profile_picture"`
}
type StoreProfileResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"store_name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	StoreImage  string `json:"store_image"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// WishlistResponse represents the response for the wishlist API
type WishlistResponse struct {
	Status   bool   `json:"status"`
	Message  string `json:"message"`
	Wishlist []ProductResponse
}

type SalesReportSeller struct {
	TotalSalesCount  int64     `json:"total_sales_count"`
	TotalOrderAmount float64 `json:"total_order_amount"`
	TotalDiscounts   float64 `json:"total_discounts"`
	PendingCount     int64     `json:"pending_count"`
	ReturnedCount    int64     `json:"returned_count"`
	CompletedCount   int64     `json:"completed_count"`
	
}


