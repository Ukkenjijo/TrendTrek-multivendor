Here's a detailed **README.md** file for your **TrendTrek API** project:

---

# **TrendTrek API**
### **E-commerce Backend in Go**

TrendTrek API is a RESTful API designed to power an e-commerce platform. It supports functionalities for user management, seller management, product management, order handling, and more, making it a robust backend for an online marketplace.

---

## **Features**
- **User Management**:
  - Sign-up and login (OTP-based).
  - Profile management.
  - Wishlist and cart functionality.
  - Address management.
  - Password recovery.

- **Seller Management**:
  - Product addition and inventory control.
  - Offer creation and sales reports.
  - Order management.

- **Admin Features**:
  - Manage users, orders, and products.
  - Generate detailed sales reports (PDF & charts).
  - Analytics: top products, categories, sellers.

- **Order Handling**:
  - Add to cart, checkout, and payment integration (Razorpay).
  - Supports refunds and wallet transactions for cancellations.

---

## **Table of Contents**
1. [Technologies Used](#technologies-used)
2. [Setup Instructions](#setup-instructions)
3. [API Endpoints](#api-endpoints)
4. [Environment Variables](#environment-variables)
5. [Testing](#testing)
6. [Deployment](#deployment)
7. [License](#license)

---

## **Technologies Used**
- **Language**: Go
- **Framework**: Fiber
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Payment Gateway**: Razorpay
- **Charts**: go-chart
- **PDF Generation**: gofpdf
- **CI/CD**: GitHub Actions

---

## **Setup Instructions**

### Prerequisites
1. Install [Go](https://golang.org/dl/).
2. Install PostgreSQL.
3. Install Git.


### Clone the Repository
```bash
git clone https://github.com/yourusername/TrendTrek.git
cd TrendTrek
```

### Install Dependencies
```bash
go mod tidy
```

### Configure Environment Variables
Create a `.env` file in the root directory with the following:
```env
DB_HOST=localhost
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=trendtrek
DB_PORT=5432

JWT_SECRET_KEY=your_jwt_secret
RAZORPAY_KEY_ID=your_razorpay_key_id
RAZORPAY_SECRET_KEY=your_razorpay_secret_key
APP_PORT=3000
```

### Migrate Database
In this project gorm is used it automigrates when the program runs:


### Start the Server
```bash
go run main.go
```

The server will run at `http://localhost:3000`.

---

## **API Endpoints**

### **User Routes**
| Method | Endpoint                          | Description                  |
|--------|-----------------------------------|------------------------------|
| POST   | `/api/v1/auth/signup`             | User sign-up (OTP-based).    |
| POST   | `/api/v1/auth/login`              | Login with email/password.   |
| POST   | `/api/v1/auth/logout`             | Logout and blacklist JWT.    |
| POST   | `/api/v1/auth/forgot-password`    | Initiate password recovery.  |
| POST   | `/api/v1/auth/reset-password`     | Reset password via OTP.      |
| GET    | `/api/v1/user/profile`            | Get user profile.            |
| PUT    | `/api/v1/user/profile`            | Update user profile.         |

### **Product Routes**
| Method | Endpoint                          | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | `/api/v1/products`               | Get all products.            |
| GET    | `/api/v1/products/:id`           | Get product details.         |
| POST   | `/api/v1/seller/products`        | Add a new product (seller).  |
| PUT    | `/api/v1/seller/products/:id`    | Update product (seller).     |
| DELETE | `/api/v1/seller/products/:id`    | Soft delete product.         |

### **Order Routes**
| Method | Endpoint                          | Description                  |
|--------|-----------------------------------|------------------------------|
| POST   | `/api/v1/user/orders`            | Place an order.              |
| GET    | `/api/v1/user/orders`            | List user orders.            |
| POST   | `/api/v1/user/orders/:id/cancel` | Cancel an order item.        |
| POST   | `/api/v1/user/orders/:id/return` | Return an order item.        |

### **Admin Routes**
| Method | Endpoint                          | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | `/api/v1/admin/sales-report`     | Generate sales report.       |
| GET    | `/api/v1/admin/top-products`     | View top 10 products.        |
| GET    | `/api/v1/admin/top-sellers`      | View top 10 sellers.         |

---

## **Environment Variables**
| Variable Name           | Description                          |
|-------------------------|--------------------------------------|
| `DB_HOST`               | Database host.                      |
| `DB_USER`               | Database user.                      |
| `DB_PASSWORD`           | Database password.                  |
| `DB_NAME`               | Database name.                      |
| `DB_PORT`               | Database port (default: 5432).      |
| `JWT_SECRET_KEY`        | JWT secret key for token signing.   |
| `RAZORPAY_KEY_ID`       | Razorpay API key ID.                |
| `RAZORPAY_SECRET_KEY`   | Razorpay secret key.                |
| `APP_PORT`              | Application port (default: 3000).   |

---



## **Deployment**



### **Manual Deployment**
1. Copy the binary and `.env` file to your server.
2. Start the service with `systemd` or a process manager like `PM2`.

---

## **License**
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

This README serves as a comprehensive guide for anyone working on or deploying the TrendTrek API. Let me know if you'd like to customize or add any sections!