package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	//err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *UserDAO) Insert(ctx context.Context, u *User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now

	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) UpdateUserProfile(ctx context.Context, userID int64, nickname string, birthdayStr string, bio string) error {
	now := time.Now().UnixMilli()

	// 将 birthdayStr 转换为 time.Time 类型
	birthday, err := time.Parse("2006-01-02", birthdayStr)
	if err != nil {
		return err
	}

	birthdayPointer := &birthday // 将 birthday 转换为 *time.Time

	return dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(User{
		Nickname: nickname,
		Birthday: birthdayPointer, // 使用 *time.Time 类型
		Bio:      bio,
		Utime:    now,
	}).Error
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"` // 全部用户唯一
	Email    string `gorm:"unique"`
	Password string
	Nickname string     `gorm:"size:100"`  // 限制昵称长度为100
	Birthday *time.Time `gorm:"type:date"` // 使用日期类型
	Bio      string     `gorm:"type:text"` // 使用文本类型存储较长文本
	Ctime    int64      // 创建时间，毫秒数
	Utime    int64      // 更新时间，毫秒数
}

func (dao *UserDAO) FindByID(id int64) (User, error) {
	var u User
	err := dao.db.First(&u, id).Error
	return u, err
}

func (dao *UserDAO) Update(user User) error {
	return dao.db.Save(&user).Error
}
