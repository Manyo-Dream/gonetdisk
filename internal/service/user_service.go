package service

import (
	"errors"
	"fmt"

	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo   *repository.UserRepo
	jwtManager *util.JWTManager
}

func NewUserService(userRepo *repository.UserRepo, jwtManger *util.JWTManager) *UserService {
	return &UserService{userRepo: userRepo, jwtManager: jwtManger}
}

func (s *UserService) Register(email, username, password string) (*dto.RegisterResponse, error) {
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, Conflict("邮箱地址已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, Internal(fmt.Sprintf("检查邮箱是否已存在失败: %s", err))
	}

	_, err = s.userRepo.GetByUserName(username)
	if err == nil {
		return nil, Conflict("用户名已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, Internal(fmt.Sprintf("检查用户名是否已存在失败: %s", err))
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal(fmt.Sprintf("密码加密失败: %s", err))
	}

	user := &model.User{
		Username:      username,
		Email:         email,
		Password_Hash: string(hashPassword),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, Internal(fmt.Sprintf("用户创建失败: %s", err))
	}

	return &dto.RegisterResponse{
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}, nil
}

func (s *UserService) Login(email, password string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("用户不存在")
	}
	if err != nil {
		return nil, Internal(fmt.Sprintf("查询用户失败: %s", err))
	}
	if user.Status == 1 {
		return nil, Forbidden("用户已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(password)); err != nil {
		return nil, Unauthorized("密码错误")
	}

	// JWT 生成 token
	token, err := s.jwtManager.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, user.Email)
	if err != nil {
		return nil, Internal(fmt.Sprintf("JWT 生成失败: %s", err))
	}

	return &dto.LoginResponse{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
	}, nil
}

func (s *UserService) GetUserInfo(email string) (*dto.UserInfoGetResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("用户不存在")
	}
	if err != nil {
		return nil, Internal(fmt.Sprintf("查询用户失败: %s", err))
	}

	return &dto.UserInfoGetResponse{
		Email:     user.Email,
		Username:  user.Username,
		AvatarUrl: user.Avatar_Url,
	}, nil
}

func (s *UserService) UpdateUserInfo(userID uint64, username, avatarUrl *string) (*dto.UserInfoUpdateResponse, error) {
	updates := make(map[string]any)

	if username != nil {
		updates["username"] = *username
	}

	if avatarUrl != nil {
		updates["avatar_url"] = *avatarUrl
	}

	if len(updates) == 0 {
		return nil, BadRequest("没有更新数据")
	}

	user, err := s.userRepo.UserInfoUpdate(userID, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("用户不存在")
		}
		return nil, Internal(fmt.Sprintf("更新用户信息失败: %s", err))
	}

	return &dto.UserInfoUpdateResponse{
		Email:     user.Email,
		Username:  user.Username,
		AvatarUrl: user.Avatar_Url,
	}, nil
}
