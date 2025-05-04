package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sf-finances/src/config"
	"sf-finances/src/types"
	"sf-finances/src/models"
	"sf-finances/src/repository"
)

var (
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrUserExists         = errors.New("пользователь уже существует")
)

type UserService interface {
	Register(ctx context.Context, req types.RegisterReq) (int64, error)
	Login(ctx context.Context, req types.LoginReq) (string, error)
	ParseToken(tokenString string) (int64, error)
}

type userService struct {
	userRepo repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, jwtCfg config.JWTConfig) UserService {
	return &userService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *userService) Register(ctx context.Context, req types.RegisterReq) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *userService) Login(ctx context.Context, req types.LoginReq) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                                   
		"exp": time.Now().Add(s.jwtCfg.ExpiresIn).Unix(), 
		"iat": time.Now().Unix(),                         
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи токена")
		}
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("невалидный токен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("невалидные claims")
	}

	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("невалидный ID пользователя")
	}

	return int64(userID), nil
}