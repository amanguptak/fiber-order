# Authentication Implementation Plan

## Goal Description
Implement secure Authentication (Register, Login, Logout, Refresh Token) functionality using JWT. This plan includes **complete, copy-paste ready code** with detailed explanations.

## User Review Required
> [!IMPORTANT]
> **Dependencies**: We will install `golang.org/x/crypto/bcrypt` (for password hashing) and `github.com/golang-jwt/jwt/v5` (for tokens). We already have `github.com/go-playground/validator/v10` from previous steps.

## Step-by-Step Implementation

### Step 1: Install Dependencies
Run the following command in your terminal:
```bash
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

### Step 2: Update Database Model
We need to store the user's email and their *hashed* password.

#### [MODIFY] [models/user.go](file:///c:/Projects/golang/go-fiber/models/user.go)
```go
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    FirstName string    `json:"firstName"`
    LastName  string    `json:"lastName"`
    
    // [NEW] Email must be unique so two users can't have the same email.
    // gorm:"unique" tells the database to create a Unique Index.
    Email     string    `json:"email" gorm:"unique"` 

    // [NEW] We store the HASHED password here, not the real one.
    // json:"-" is CRITICAL. It tells Go: "Never send this field in the JSON response".
    // This prevents us from accidentally leaking password hashes to the frontend.
    Password  []byte    `json:"-"`
}
```

### Step 3: Create Auth DTOs & Response DTOs
Define what data the user sends for Register and Login, and what we send back.

#### [NEW] [dtos/auth_dto.go](file:///c:/Projects/golang/go-fiber/dtos/auth_dto.go)
```go
package dtos

import "github.com/google/uuid"

// Request DTOs
type RegisterRequest struct {
    FirstName string `json:"firstName" validate:"required"`
    LastName  string `json:"lastName" validate:"required"`
    Email     string `json:"email" validate:"required,email"` // Validate it's a real email format
    Password  string `json:"password" validate:"required,min=6"` // Enforce min length for security
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// Response DTOs (Safer than returning models directly)
type UserResponse struct {
    ID        uuid.UUID `json:"id"`
    FirstName string    `json:"firstName"`
    LastName  string    `json:"lastName"`
    Email     string    `json:"email"`
}
```

### Step 4: Create JWT Helper (Best Practice)
Instead of writing JWT logic inside every route, we create a reusable helper function. This keeps our code clean and DRY (Don't Repeat Yourself).

#### [NEW] [helpers/jwt_helper.go](file:///c:/Projects/golang/go-fiber/helpers/jwt_helper.go)
```go
package helpers

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// SecretKey is the private key used to sign tokens. 
// If a hacker gets this, they can forge tokens and impersonate anyone.
const SecretKey = "secret" 

// GenerateToken creates a new JWT token for a user
// issuer: The User ID (who owns this token)
// expirationTime: When this token should stop working
func GenerateToken(issuer string, expirationTime time.Time) (string, error) {
    // jwt.NewWithClaims creates a new Token object.
    // jwt.SigningMethodHS256: The algorithm used to sign the token (HMAC SHA256).
    // jwt.MapClaims: A map to hold the data (payload) we want to store inside the token.
    claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "iss": issuer,           // "iss" (Issuer) is a standard JWT claim for the User ID.
        "exp": expirationTime.Unix(), // "exp" (Expiration) tells the browser/server when it expires.
    })

    // SignedString takes our SecretKey and uses it to cryptographically sign the token.
    // This generates the final "ey..." string you see.
    return claims.SignedString([]byte(SecretKey))
}

// ParseToken validates the token and returns the claims
func ParseToken(tokenString string) (*jwt.Token, error) {
    // jwt.ParseWithClaims does 3 things:
    // 1. Decodes the token string.
    // 2. Checks the signature using the function we provide (to make sure it wasn't fake).
    // 3. Checks standard claims like "exp" (expiration) automatically.
    return jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
        // We must return the SAME SecretKey we used to sign it.
        return []byte(SecretKey), nil
    })
}
```

### Step 5: Create Auth Middleware (Best Practice)
Middleware protects routes. It checks if the user is logged in *before* the request reaches the controller.

#### [NEW] [middleware/auth_middleware.go](file:///c:/Projects/golang/go-fiber/middleware/auth_middleware.go)
```go
package middleware

import (
    "github.com/amanguptak/fiber-api/helpers"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
)

// IsAuthenticated is a middleware function that runs before the main controller.
func IsAuthenticated(c *fiber.Ctx) error {
    // 1. Get Token from Cookie
    // We expect the token to be in a cookie named "jwt".
    cookie := c.Cookies("jwt") 

    // 2. Parse Token using our helper
    // This checks if the token is real, not expired, and signed by us.
    token, err := helpers.ParseToken(cookie)

    // 3. Check Validity
    // If there was an error parsing (err != nil) OR the token is marked invalid...
    if err != nil || !token.Valid {
        // ...return a 401 Unauthorized error and stop the request here.
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "message": "unauthenticated",
        })
    }
    
    // 4. Pass to next handler
    // If everything is good, allow the request to go to the next function (the controller).
    return c.Next()
}
```

### Step 6: Implement Auth Logic (Using Helpers & Validation)
Now our controller is much cleaner because it uses the helpers.

#### [NEW] [routes/auth.go](file:///c:/Projects/golang/go-fiber/routes/auth.go)

**1. Register Function**
```go
func Register(c *fiber.Ctx) error {
    var data dtos.RegisterRequest
    // Parse the JSON body into our struct
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
    }

    // [NEW] Validation Step
    // We use the validator library to check rules like "required", "email", "min=6".
    // If validation fails, we return a 400 error immediately.
    validate := validator.New()
    if err := validate.Struct(data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // BEHIND THE SCENE: bcrypt.GenerateFromPassword
    // 1. It adds a random "Salt" to the password (so "password123" looks different for every user).
    // 2. It runs the hashing algorithm 14 times (Cost=14). This makes it slow on purpose,
    //    so hackers can't try millions of passwords per second.
    password, _ := bcrypt.GenerateFromPassword([]byte(data.Password), 14)

    user := models.User{
        FirstName: data.FirstName,
        LastName:  data.LastName,
        Email:     data.Email,
        Password:  password, // Save the HASH, not the plain text
    }

    // Save to Database. If email exists, GORM returns an error because of `gorm:"unique"`.
    if err := helpers.DB().Create(&user).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Could not create user"})
    }

    // [NEW] Return Response DTO
    // Instead of returning the raw 'user' model, we create a clean 'UserResponse' DTO.
    // This ensures we only send back the fields we explicitly want to expose.
    response := dtos.UserResponse{
        ID:        user.ID,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Email:     user.Email,
    }

    return c.JSON(response)
}
```

**2. Login Function (Using Helper)**
```go
func Login(c *fiber.Ctx) error {
    var data dtos.LoginRequest
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
    }

    // [NEW] Validation Step
    // Even for login, we should validate that an email and password were actually sent.
    validate := validator.New()
    if err := validate.Struct(data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    var user models.User
    // Find the user with this email. 
    // .First() adds "LIMIT 1" to the SQL query for efficiency.
    helpers.DB().Where("email = ?", data.Email).First(&user)

    // If ID is empty (uuid.Nil), it means no user was found.
    if user.ID == uuid.Nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
    }

    // BEHIND THE SCENE: bcrypt.CompareHashAndPassword
    // It takes the stored hash (user.Password) and the input password (data.Password).
    // It re-hashes the input using the SAME salt from the stored hash and checks if they match.
    if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data.Password)); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Incorrect password"})
    }

    // USE HELPER: Generate Access Token (15 mins)
    // We use our helper to create a short-lived token for API access.
    token, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Minute * 15))

    // USE HELPER: Generate Refresh Token (7 days)
    // We create a long-lived token so the user doesn't have to login every 15 mins.
    refreshToken, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Hour * 24 * 7))

    // Set Cookie
    // We put the Refresh Token in an HttpOnly cookie.
    cookie := fiber.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Expires:  time.Now().Add(time.Hour * 24 * 7),
        HTTPOnly: true, // CRITICAL: JavaScript cannot read this. Prevents XSS attacks.
    }
    c.Cookie(&cookie)

    return c.JSON(fiber.Map{
        "message": "success",
        "token": token, // Send Access Token in JSON (Frontend attaches it to Headers)
    })
}
```

**3. Logout Function**
```go
func Logout(c *fiber.Ctx) error {
    // To "delete" a cookie, we overwrite it with an empty value 
    // and set the expiration date to the past (1 hour ago).
    // The browser sees it's expired and removes it immediately.
    cookie := fiber.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Expires:  time.Now().Add(-time.Hour),
        HTTPOnly: true,
    }
    c.Cookie(&cookie)
    return c.JSON(fiber.Map{"message": "success"})
}
```

**4. Refresh Function**
```go
func Refresh(c *fiber.Ctx) error {
    // Get the cookie automatically sent by the browser
    cookie := c.Cookies("refresh_token")

    // USE HELPER: Parse Token
    // Validate that the refresh token is valid and not expired.
    token, err := helpers.ParseToken(cookie)

    if err != nil || !token.Valid {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthenticated"})
    }

    // Get the data (Claims) from the token
    claims := token.Claims.(*jwt.MapClaims)

    // USE HELPER: Generate NEW Access Token
    // Since the refresh token is valid, we issue a brand new Access Token (15 mins).
    newToken, _ := helpers.GenerateToken(claims["iss"].(string), time.Now().Add(time.Minute * 15))

    return c.JSON(fiber.Map{
        "token": newToken,
    })
}
```

### Step 7: Register Routes
Connect the new functions to URLs.

#### [MODIFY] [main.go](file:///c:/Projects/golang/go-fiber/main.go)
```go
func main() {
    // ... existing setup ...
    app.Post("/api/register", routes.Register)
    app.Post("/api/login", routes.Login)
    app.Post("/api/logout", routes.Logout)
    app.Post("/api/refresh", routes.Refresh) 
    // ... existing routes ...
}
```

### Step 8: Cleanup Legacy Code
Since `Register` now handles user creation with security, the old `CreateUser` function is redundant.

**1. Remove Route from `main.go`**
Find and delete: `app.Post("/api/users", routes.CreateUser)`

**2. Remove Function from `routes/user.go`**
Delete the entire `CreateUser` function.

## Verification Plan
1. **Login**: Get `token` (JSON) and `refresh_token` (Cookie).
2. **Refresh**: Call `/api/refresh` (browser sends cookie automatically). Expect new `token`.
