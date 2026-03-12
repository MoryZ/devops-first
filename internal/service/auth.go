package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"devops-first/internal/database"
	"devops-first/internal/model"
)

const (
	TokenExpiration = 7 * 24 * time.Hour // 7 days
	JWTSecret       = "your-secret-key"  // Should be from environment in production
)

type AuthService struct {
	db *gorm.DB
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"`
	Remark   string `json:"remark"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Remark   string `json:"remark"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

// Register creates a new user
func (s *AuthService) Register(req RegisterRequest) (*TokenResponse, error) {
	// Check if user exists
	var existing model.User
	if err := s.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Remark:   req.Remark,
		Active:   true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	return s.generateToken(&user)
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(req LoginRequest) (*TokenResponse, error) {
	// Find user
	var user model.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.New("user is not active")
	}

	// Generate token
	return s.generateToken(&user)
}

// generateToken creates a JWT token for a user
func (s *AuthService) generateToken(user *model.User) (*TokenResponse, error) {
	expiresAt := time.Now().Add(TokenExpiration)

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &TokenResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Remark:   user.Remark,
		},
	}, nil
}

// VerifyToken parses and validates a JWT token
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
