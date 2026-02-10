package service

import (
	"errors"
	"fmt"

	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   *repository.UserRepo
	jwtManager *util.JWTManager
}

func NewUserService(userRepo *repository.UserRepo, jwtManger *util.JWTManager) *UserService {
	return &UserService{userRepo: userRepo, jwtManager: jwtManger}
}

func (s *UserService) Register(email, username, password string) (*dto.RegisterResponse, error) {
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("邮箱地址已存在")
	}

	existingUser, _ = s.userRepo.GetByUserName(username)
	if existingUser != nil {
		return nil, errors.New("用户名称已存在")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	user := &model.User{
		Username:      username,
		Email:         email,
		Password_Hash: string(hashPassword),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("用户创建失败: %w", err)
	}

	return &dto.RegisterResponse{
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}, nil

}

func (s *UserService) Login(email, password string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	// JWT 生成 token
	token, err := s.jwtManager.GenerateToken(user.Username, user.Email)
	if err != nil {
		return nil, errors.New("JWT 生成失败")
	}

	return &dto.LoginResponse{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
	}, nil
}

func (s *UserService) GetUserInfo(email string) (*dto.UserInfoGetResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	return &dto.UserInfoGetResponse{
		Email:     user.Email,
		Username:  user.Username,
		AvatarUrl: user.Avatar_Url,
	}, nil
}

func (s *UserService) UpdateUserInfo(email, username, avatarUrl *string) (*dto.UserInfoUpdateResponse, error) {
	updates := make(map[string]any)

	if email != nil {
		updates["email"] = *email
	}

	if username != nil {
		updates["username"] = *username
	}

	if avatarUrl != nil {
		updates["avatar_url"] = *avatarUrl
	}

	if len(updates) == 0 {
		return nil, errors.New("没有更新数据")
	}

	user, err := s.userRepo.UserInfoUpdate(email, updates)
	if err != nil {
		return nil, errors.New("更新用户信息失败")
	}

	return &dto.UserInfoUpdateResponse{
		Email:     user.Email,
		Username:  user.Username,
		AvatarUrl: user.Avatar_Url,
	}, nil
}
