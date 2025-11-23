# Refresh Token Rotation Implementation Plan (Advanced Security)

## Goal Description
Upgrade the authentication system to use **Refresh Token Rotation**. This involves storing refresh tokens in the database, revoking them upon use, and detecting token reuse (theft).

## Why This is "Best Practice"
1.  **Revocation:** You can ban a user instantly by deleting their token from the DB.
2.  **Theft Detection:** If a hacker steals a token and uses it, the real user will also try to use it later. The system detects this "double use" and locks the account.

## Step-by-Step Implementation

### Step 1: Create Refresh Token Model
We need a table to track every refresh token issued. We use UUIDs for consistency.

#### [NEW] [models/refresh_token.go](file:///c:/Projects/golang/go-fiber/models/refresh_token.go)
```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type RefreshToken struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    
    // Link to the User
    UserID    uuid.UUID `gorm:"type:uuid;index"` 
    User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    
    // Security Fields
    TokenHash string    `gorm:"not null;index"` // We store the SHA256 hash, not the raw token
    IsRevoked bool      `gorm:"default:false"`  // True if this token has been used/killed
    ExpiresAt time.Time `gorm:"not null"`
    
    CreatedAt time.Time
}

// BeforeCreate generates a new UUID before saving to the DB
func (token *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
    token.ID = uuid.New()
    return
}
```

### Step 2: Update Database Connection
Register the new model in your database setup so GORM creates the table.

#### [MODIFY] [database/database.go](file:///c:/Projects/golang/go-fiber/database/database.go)

**Before:**
```go
err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{})
```

**After:**
```go
// Add &models.RefreshToken{} to the list
err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.RefreshToken{})
```

### Step 3: Create Token Repository (Helper)
Create a helper to handle the complex DB logic for rotation.

#### [NEW] [repositories/token_repository.go](file:///c:/Projects/golang/go-fiber/repositories/token_repository.go)
```go
package repositories

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
    "github.com/amanguptak/fiber-api/helpers"
    "github.com/amanguptak/fiber-api/models"
    "github.com/google/uuid"
)

// HashToken converts the raw JWT string into a secure hash for storage
func HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}

// StoreRefreshToken saves a new token to the DB
func StoreRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error {
    refreshToken := models.RefreshToken{
        UserID:    userID,
        TokenHash: HashToken(token),
        ExpiresAt: expiresAt,
    }
    return helpers.DB().Create(&refreshToken).Error
}

// RotateRefreshToken handles the security logic
// Returns: (NewTokenString, Error)
func RotateRefreshToken(oldTokenString string) (string, error) {
    tx := helpers.DB().Begin() // Start Transaction
    
    // 1. Find the old token
    var dbToken models.RefreshToken
    hash := HashToken(oldTokenString)
    if err := tx.Where("token_hash = ?", hash).First(&dbToken).Error; err != nil {
        tx.Rollback()
        return "", err // Token not found
    }

    // 2. Reuse Detection (Theft!)
    if dbToken.IsRevoked {
        // ALARM: This token was already used!
        // Security Policy: Revoke ALL tokens for this user (Force Logout)
        tx.Model(&models.RefreshToken{}).Where("user_id = ?", dbToken.UserID).Update("is_revoked", true)
        tx.Commit()
        return "", fmt.Errorf("token reuse detected")
    }

    // 3. Revoke the old token
    dbToken.IsRevoked = true
    tx.Save(&dbToken)

    // 4. Issue New Token
    newToken, err := helpers.GenerateToken(dbToken.UserID.String(), time.Now().Add(7*24*time.Hour))
    if err != nil {
        tx.Rollback()
        return "", err
    }
    
    // 5. Save New Token
    newDbToken := models.RefreshToken{
        UserID:    dbToken.UserID,
        TokenHash: HashToken(newToken),
        ExpiresAt: time.Now().Add(7*24*time.Hour),
    }
    if err := tx.Create(&newDbToken).Error; err != nil {
        tx.Rollback()
        return "", err
    }

    tx.Commit() // Save everything
    return newToken, nil
}
```

### Step 4: Update Auth Routes
Modify `Login` and `Refresh` to use the new repository.

#### [MODIFY] [routes/auth.go](file:///c:/Projects/golang/go-fiber/routes/auth.go)

**Update Login Function:**

**Before:**
```go
	// USE HELPER: Generate Refresh Token (7 days)
	// We create a long-lived token so the user doesn't have to login every 15 mins.
	refreshToken, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Hour*24*7))
	// Set Cookie
```

**After:**
```go
	// USE HELPER: Generate Refresh Token (7 days)
	// We create a long-lived token so the user doesn't have to login every 15 mins.
	refreshToken, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Hour*24*7))
	
    // [NEW] Store Refresh Token in DB
    // We log error but don't fail login, though in strict mode you might want to fail.
    if err := repositories.StoreRefreshToken(user.ID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save session"})
    }

	// Set Cookie
```

**Update Refresh Function (To be implemented):**

**Before (Stateless Plan):**
```go
    // USE HELPER: Parse Token
    token, err := helpers.ParseToken(cookie)
    // ...
    // USE HELPER: Generate NEW Access Token
    newToken, _ := helpers.GenerateToken(...)
```

**After (Rotation Plan):**
```go
    // [NEW] Rotate Logic (Replaces simple ParseToken)
    newToken, err := repositories.RotateRefreshToken(oldToken)
    if err != nil {
        // If reuse detected or invalid, clear cookie and error out
        c.ClearCookie("refresh_token")
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid token"})
    }
    // Return newToken...
```

### Step 5: Cleanup (Cron Job)
Since we store every token, the table will grow forever. We need a background job to delete expired tokens.

*   **Recommendation:** Run a daily job: `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
