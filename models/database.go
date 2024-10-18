package models

import (
	"log"

	"gorm.io/gorm"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleSeller   Role = "seller"
	RoleAdmin    Role = "admin"
)

type User struct {
	gorm.Model
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	Name           string    `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string    `json:"email" gorm:"unique;not null"`
	PhoneNumber    string    `json:"phone" gorm:"unique;not null"`
	Blocked        bool      `gorm:"column:blocked;type:bool" json:"blocked"`
	Role           Role      `gorm:"column:role;type:varchar(50);default:'customer'" json:"role"`
	HashedPassword string    `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Verified       bool      `json:"verified" gorm:"default:false"`
	IsAdmin        bool      `gorm:"default:false"`
	ProfilePicture string    `gorm:"size:255"`
	UserCart       *Cart     `gorm:"foreignKey:UserID" json:"user_cart"`
	Addresses      []Address `json:"addresses" gorm:"foreignKey:UserID"`
}

type Address struct {
	gorm.Model
	UserID    uint   `json:"user_id"`
	Street    string `json:"street" validate:"required"`
	City      string `json:"city" validate:"required"`
	State     string `json:"state" validate:"required"`
	Country   string `json:"country" validate:"required"`
	ZipCode   string `json:"zip_code" validate:"required"`
	IsDefault bool   `json:"is_default" gorm:"default:false"`
}

type Store struct {
	gorm.Model
	Name        string `json:"store_name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	StoreImage  string `json:"store_image"` // Added StoreImage field
	Certificate string `json:"certificate"`
	UserID      uint
	User        *User `gorm:"references:ID;foreignKey:UserID"`
}

type Category struct {
	gorm.Model
	Name             string    `gorm:"type:varchar(100);not null" json:"name"`
	ParentCategoryID *uint     `gorm:"index;null" json:"parent_category_id"`
	ParentCategory   *Category `gorm:"foreignKey:ParentCategoryID" json:"parent_category,omitempty"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
}
type Product struct {
	gorm.Model
	StoreID       uint     `gorm:"not null" json:"store_id"`                                          // Foreign key referencing Store
	Store         *Store   `gorm:"foreignKey:StoreID" json:"store"`                                   // Relation to Store model
	Name          string   `gorm:"type:varchar(100);not null" json:"name"`                            // Product name
	Description   string   `gorm:"type:text" json:"description,omitempty"`                            // Product description (optional)
	Price         float64  `gorm:"type:decimal(10,2);not null" json:"price" validate:"required,gt=0"` // Product price with 2 decimal places
	StockQuantity int      `gorm:"default:0;check:stock_quantity >= 0" json:"stock_quantity"`         // Stock quantity (default 0)
	CategoryID    uint     `gorm:"not null" json:"category_id"`                                       // Foreign key referencing Category
	Category      Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`                   // Relation to Category model
	IsActive      bool     `gorm:"default:true" json:"is_active"`
	Images        []Image  `gorm:"foreignKey:ProductID" json:"images,omitempty"` // Is product active (default: true)
	StockLeft     int      `gorm:"default:0" json:"stock_left"`                  // Stock left
}

func (p *Product) AfterUpdate(tx *gorm.DB) error {
	if p.StockLeft <= 0 {
		log.Println("Stock left:", p.StockLeft)
		if p.IsActive {
			p.IsActive = false
			// Add a check to see if the record has already been updated
			if tx.Model(p).Where("is_active = ?", false).Take(&p).Error == nil {
				log.Println("Record already updated, skipping")
				return nil
			}
			return tx.Save(p).Error
		}
	}
	return nil
}

type Cart struct {
	gorm.Model
	UserID uint       `json:"user_id"`
	Items  []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

type CartItem struct {
	gorm.Model
	CartID     uint    `json:"cart_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`       // Unit price of the product
	TotalPrice float64 `json:"total_price"` // Calculated as Quantity * Price
}

type Order struct {
	gorm.Model
	UserID      uint        `json:"user_id"`
	AddressID   uint        `json:"address_id"` // Foreign key to the selected address
	TotalAmount float64     `json:"total_amount"`
	PaymentMode string      `json:"payment_mode"` // e.g., "COD"
	Status      string      `json:"status"`       // e.g., "pending", "shipped", "delivered", "canceled"
	Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID"`

	// Address snapshot fields
	ShippingStreet  string `json:"shipping_street"`
	ShippingCity    string `json:"shipping_city"`
	ShippingState   string `json:"shipping_state"`
	ShippingCountry string `json:"shipping_country"`
	ShippingZipCode string `json:"shipping_zip_code"`
}
type OrderItem struct {
	gorm.Model
	OrderID    uint    `json:"order_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	Status     string  `json:"status" gorm:"default:'pending'"` // individual item status
	TotalPrice float64 `json:"total_price"`                     // Price * Quantity
}

type Image struct {
	gorm.Model
	URL       string `gorm:"type:varchar(255);not null" json:"url"` // URL or path of the image
	ProductID *uint  `gorm:"index" json:"product_id,omitempty"`     // Optional: Foreign key to Product (nullable)

}
