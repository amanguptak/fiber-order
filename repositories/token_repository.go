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

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])

}

func StoreRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	refreshToken := models.RefreshToken{
		UserID:    userID,
		TokenHash: HashToken(token),
		ExpiresAt: expiresAt,
	}

	return helpers.DB().Create(&refreshToken).Error
}

func RotateRefreshToken(oldTokenString string) (string, string, error) {
	tx := helpers.DB().Begin()
	var dbToken models.RefreshToken
	hash := HashToken(oldTokenString)
	if err := tx.Where("token_hash = ?", hash).First(&dbToken).Error; err != nil {
		tx.Rollback()
		return "", "", err // ✅ Return 2 empty strings
	}

	if dbToken.IsRevoked {
		tx.Model(&models.RefreshToken{}).Where("user_id=?", dbToken.UserID).Update("is_revoked", true)
		tx.Commit()
		return "", "", fmt.Errorf("token reuse detected")
	}

	dbToken.IsRevoked = true
	tx.Save(&dbToken)
	// ✅ Generate NEW Access Token (15 mins)
	newAccessToken, err := helpers.GenerateToken(dbToken.UserID.String(), time.Now().Add(15*time.Minute))
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	// ✅ Generate NEW Refresh Token (7 days)
	newRefreshToken, err := helpers.GenerateToken(dbToken.UserID.String(), time.Now().Add(7*24*time.Hour))
	if err != nil {
		tx.Rollback()
		return "", "", err
	}
	newDbToken := models.RefreshToken{
		UserID:    dbToken.UserID,
		TokenHash: HashToken(newRefreshToken), // ✅ Hash the REFRESH token
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := tx.Create(&newDbToken).Error; err != nil {
		tx.Rollback()
		return "", "", err 
	}
	tx.Commit() // Save everything
	return newAccessToken, newRefreshToken, nil
}

/// ********** explaination of above function line by line  in token_repository.md---------
