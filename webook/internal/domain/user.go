package domain

import (
	"gorm.io/gorm"
	"time"
)

// User 领域对象，是 DDD 中的 entity
// BO(business object)
type User struct {
	Id       int64      `gorm:"primaryKey,autoIncrement"`
	Email    string     `gorm:"unique"`
	Password string     `json:"-"`
	Nickname string     `gorm:"size:100"`  // 限制昵称长度为100
	Birthday *time.Time `gorm:"type:date"` // 使用日期类型
	Bio      string     `gorm:"type:text"` // 使用文本类型存储较长文本
	Ctime    int64      // 创建时间，毫秒数
	Utime    int64      // 更新时间，毫秒数
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Birthday == nil || u.Birthday.IsZero() {
		u.Birthday = nil
	}
	return
}

//type Address struct {
//}
