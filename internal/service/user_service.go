package service

import (
	"errors"
	"fmt"

	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Register(email, username, password string) error {
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return errors.New("邮箱地址已存在")
	}

	existingUser, _ = s.userRepo.GetByUserName(username)
	if existingUser != nil {
		return errors.New("用户名称已存在")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user := &model.User{
		UserName:      username,
		Email:         email,
		Password_Hash: string(hashPassword),
	}

	return s.userRepo.Create(user)
}

func (s *UserService) Login(email, password string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != 0 {
		return nil, errors.New("用户已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	// JWT 生成 token
	
}
