package service

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService struct {
	repo    *repository.UserRepository
	userDAO *dao.UserDAO
}

type UserEditRequest struct {
	Nickname string `json:"nickname" binding:"required,max=100"`
	Birthday string `json:"birthday" binding:"required,datetime=2006-01-02"`
	Bio      string `json:"bio" binding:"max=500"`
}

func NewUserService(repo *repository.UserRepository, dao *dao.UserDAO) *UserService {
	return &UserService{
		repo:    repo,
		userDAO: dao,
	}
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码了
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return svc.repo.Create(ctx, u)
}

// 自定义一些可能的错误常量
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidNickname = errors.New("invalid nickname")
	ErrInvalidBirthday = errors.New("invalid birthday format")
	ErrInvalidBio      = errors.New("invalid bio")
)

func (s *UserService) UpdateUserProfile(userID int64, nickname string, birthdayStr string, bio string) error {
	// 参数验证
	if nickname == "" {
		return ErrInvalidNickname
	}

	parsedBirthday, err := time.Parse("2006-01-02", birthdayStr)
	if err != nil {
		return ErrInvalidBirthday
	}

	if len(bio) > 500 { // 假设500是bio的最大长度
		return ErrInvalidBio
	}

	// 找到要更新的用户
	user, err := s.userDAO.FindByID(userID)
	if err != nil {
		// 如果userDAO.FindByID返回的错误表示用户不存在，则返回自定义错误
		return ErrUserNotFound
	}

	// 更新用户信息
	user.Nickname = nickname

	// 将 time.Time 转换为 *time.Time
	parsedBirthdayPtr := &parsedBirthday
	user.Birthday = parsedBirthdayPtr

	user.Bio = bio

	// 保存更新后的用户信息
	err = s.userDAO.Update(user)
	if err != nil {
		// 这里可以继续进行更详细的错误处理，例如数据库错误等
		return err
	}

	return nil
}

func (s *UserService) GetUserProfile(userID int64) (dao.User, error) {
	// 直接从数据库获取用户信息
	return s.userDAO.FindByID(userID)
}

func (svc *UserService) GetUserByID(userID int64) (*domain.User, error) {
	user, err := svc.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (svc *UserService) GetUserProfileByID(userID int64) (*domain.User, error) {
	user, err := svc.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
