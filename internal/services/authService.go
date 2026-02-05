package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImp struct{
	userService UserService
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	GenerateToken(User *models.User) (string, error)
	ValidateToken(token string) (*models.User, error)
	Login(username, password string) (*TokenPair, error)
	GenerateRefreshToken(user *models.User) (string, error)
	ValidateRefreshToken(tokenString string) (*models.User, error)
	SaveRefreshToken(userId, refreshToken string) error
	ClearRefreshToken(userId string) error
	RefreshTokens(refreshToken string) (*TokenPair, error)
}

func NewAuthService() AuthService {
	return &AuthServiceImp{
		userService: NewUserService(),
	}
}

func (service *AuthServiceImp) Login(username, password string) (*TokenPair, error) {
	_, err := config.GetDB()
	if err != nil {
		return nil, fmt.Errorf("error al conectar a la base de datos: %v", err)
	}

	user, err := service.userService.GetUserByUserName(username)
	if err != nil {
		return nil, fmt.Errorf("error al buscar el usuario: %v", err)
	}

	if !CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("la contraseña no es válida")
	}

	accessToken, err := service.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("error al generar el access token: %v", err)
	}

	refreshToken, err := service.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("error al generar el refresh token: %v", err)
	}

	if err := service.SaveRefreshToken(user.Id, refreshToken); err != nil {
		return nil, fmt.Errorf("error al guardar el refresh token: %v", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (service *AuthServiceImp) GenerateToken(user *models.User) (string, error) {

	SecretToken := []byte(config.GetConfig().JWTSecretKey)

	//crear token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.Id,               // Identificador único del usuario
		"username": user.Username,         // Nombre de usuario para referencia
		"email":    user.Email,            
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Expira en 24 horas
	})

	//firmar token
	tokenString, err := token.SignedString(SecretToken)
	if err != nil {
		return "", fmt.Errorf("error al firmar el token: %v", err)
	}

	return tokenString, nil
}

func (service *AuthServiceImp) ValidateToken(tokenString string) (*models.User, error) {
	
	SecretToken := []byte(config.GetConfig().JWTSecretKey)

	// Parsear y verificar el token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return SecretToken, nil
	})

	if err != nil {
		// Error al parsear o verificar el token
		return nil, fmt.Errorf("error al parsear el token: %v", err)
	}

	// Extraer y validar los claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Validar y construir el objeto usuario
		id, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("user_id no es válido")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, fmt.Errorf("username no es válido")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("email no es válido")
		}

		user := &models.User{
			Id:       id,
			Username: username,
			Email:    email,
		}

		return user, nil
	}

	// Si el token no es válido o los claims no son correctos
	return nil, fmt.Errorf("token inválido o claims inválidos")
}

// Función para hashear una contraseña
func HashPassword(password string) (string, error) {
	// Generar el hash de la contraseña con un costo predeterminado (bcrypt.DefaultCost)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error al generar el hash: %v", err)
	}
	return string(hashedPassword), nil
}

// Función para comparar una contraseña sin hashear con su hash
func CheckPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// hashSHA256 genera un hash SHA-256 determinístico (rápido, suficiente para tokens aleatorios)
func hashSHA256(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// GenerateRefreshToken genera un JWT de refresh con exp de 7 días.
// Contiene el user_id para poder hacer lookup directo en DB.
func (service *AuthServiceImp) GenerateRefreshToken(user *models.User) (string, error) {
	secretToken := []byte(config.GetConfig().JWTSecretKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"type":    "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(secretToken)
	if err != nil {
		return "", fmt.Errorf("error signing refresh token: %v", err)
	}

	return tokenString, nil
}

// ValidateRefreshToken parsea el JWT para obtener user_id,
// busca al usuario en DB y compara el hash SHA-256 almacenado.
func (service *AuthServiceImp) ValidateRefreshToken(refreshToken string) (*models.User, error) {
	secretToken := []byte(config.GetConfig().JWTSecretKey)

	parsedToken, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretToken, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	userId, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in refresh token")
	}

	// Lookup directo por user_id (O(1), no iterar todos los usuarios)
	user, err := service.userService.GetUserByID(userId)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Comparar hash SHA-256 almacenado con el del token recibido
	if user.RefreshToken == "" || user.RefreshToken != hashSHA256(refreshToken) {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	return user, nil
}

// SaveRefreshToken guarda un hash SHA-256 del refresh token en la DB
func (service *AuthServiceImp) SaveRefreshToken(userId, refreshToken string) error {
	db, err := config.GetDB()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	hashed := hashSHA256(refreshToken)

	if err := db.Model(&models.User{}).Where("id = ?", userId).Update("refresh_token", hashed).Error; err != nil {
		return fmt.Errorf("error saving refresh token: %v", err)
	}

	return nil
}

// ClearRefreshToken limpia el refresh token del usuario (logout)
func (service *AuthServiceImp) ClearRefreshToken(userId string) error {
	db, err := config.GetDB()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	if err := db.Model(&models.User{}).Where("id = ?", userId).Update("refresh_token", "").Error; err != nil {
		return fmt.Errorf("error clearing refresh token: %v", err)
	}

	return nil
}

// RefreshTokens valida el refresh token actual y genera un nuevo par de tokens (rotation)
func (service *AuthServiceImp) RefreshTokens(refreshToken string) (*TokenPair, error) {
	user, err := service.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	accessToken, err := service.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %v", err)
	}

	newRefreshToken, err := service.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %v", err)
	}

	if err := service.SaveRefreshToken(user.Id, newRefreshToken); err != nil {
		return nil, fmt.Errorf("error saving refresh token: %v", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
