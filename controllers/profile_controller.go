package controllers

import (
	"log"
	"net/http"

	"MyForum/config"
	"MyForum/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

// GetUserProfile retrieves user profile details
func GetUserProfile(c *gin.Context) {
	// Get the logged-in user ID from the context
	userID, ok := c.Get("userID")
	if !ok {
		log.Println("User ID not found in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch user information from the database
	var user models.User
	err := config.DB.QueryRow("SELECT id, email, username, name, surname FROM users WHERE id = ?", userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.Name, &user.Surname)
	if err != nil {
		log.Println("Failed to fetch user profile:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user profile"})
		return
	}

	log.Println("User profile fetched successfully, User ID:", user.ID)
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username.String,
		"name":     user.Name.String,
		"surname":  user.Surname.String,
	})
}

// ChangePassword handles changing the user's password.
func ChangePassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch user from database
	var user models.User
	err := config.DB.QueryRow("SELECT id, password FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user password"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update user password in the database
	stmt, err := config.DB.Prepare("UPDATE users SET password = ? WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(hashedPassword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

/* RegisterProfileRoutes registers routes for user profile operations
func RegisterProfileRoutes(router *gin.Engine) {
	router.GET("/profile/:id", utils.AuthRequired(), GetUserProfile)

	router.POST("/profile/change-password", utils.AuthRequired(), ChangePassword)
	//router.GET("/profile", utils.AuthRequired(), ProfileView)
}
*/
