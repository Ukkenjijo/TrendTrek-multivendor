package models

import "gorm.io/gorm"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleSeller   Role = "seller"
	RoleAdmin    Role = "admin"
)

type User struct {
	gorm.Model
	ID             uint   `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string `json:"email" gorm:"unique;not null"`
	PhoneNumber    string `json:"phone" gorm:"unique;not null"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	Role           Role   `gorm:"column:role;type:varchar(50);default:'customer'" json:"role"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required,min=8" json:"hashed_password"`
	Verified       bool   `json:"verified" gorm:"default:false"`
	IsAdmin        bool   `gorm:"default:false"`
	ProfilePicture string `gorm:"size:255"`
}

type Store struct {
	gorm.Model
	Name        string `json:"store_name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
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
	StockLeft     int      `gorm:"default:0" json:"stock_left"` // Stock left
}
func (p *Product) AfterUpdate(tx *gorm.DB) error {
	if p.StockLeft <= 0 {
		p.IsActive = false
		return tx.Save(p).Error
	}
	return nil
}

type Image struct {
	gorm.Model
	URL       string `gorm:"type:varchar(255);not null" json:"url"` // URL or path of the image
	ProductID *uint  `gorm:"index" json:"product_id,omitempty"`     // Optional: Foreign key to Product (nullable)

}
