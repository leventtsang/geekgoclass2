package repository

import (
	"context"
	"database/sql"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
	db  *sql.DB
}

// 修改构造函数以接收数据库连接对象
func NewUserRepository(dao *dao.UserDAO, db *sql.DB) *UserRepository {
	return &UserRepository{
		dao: dao,
		db:  db,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, &dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(int64) {
	// 先从 cache 里面找
	// 再从 dao 里面找
	// 找到了回写 cache
}

func (r *UserRepository) GetUserByID(userID int64) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, email, nickname, birthday, bio
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(query, userID).Scan(&user.Id, &user.Email, &user.Nickname, &user.Birthday, &user.Bio)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user found with id %d", userID)
		}
		return nil, err
	}
	return user, nil
}
