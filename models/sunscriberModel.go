package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// table subscribe
type Subscriber struct {
	gorm.Model
	Email          string `gorm:"unique_index"` //邮箱
	VerifyState    bool   `gorm:"default:'0'"`  //验证状态
	SubscribeState bool   `gorm:"default:'1'"`  //订阅状态
	OutTime        time.Time                    //过期时间
	SecretKey      string                       // 秘钥
	Signature      string                       //签名
}

// Subscriber
func (s *Subscriber) Insert() error {
	return DB.FirstOrCreate(s, "email = ?", s.Email).Error
}

func (s *Subscriber) Update() error {
	return DB.Model(s).Update(map[string]interface{}{
		"verify_state":    s.VerifyState,
		"subscribe_state": s.SubscribeState,
		"out_time":        s.OutTime,
		"signature":       s.Signature,
		"secret_key":      s.SecretKey,
	}).Error
}

func ListSubscriber(invalid bool) ([]*Subscriber, error) {
	var subscribers []*Subscriber
	db := DB.Model(&Subscriber{})
	if invalid {
		db.Where("verify_state = ? and subscribe_state = ?", true, true)
	}
	err := db.Find(&subscribers).Error
	return subscribers, err
}

func CountSubscriber() (int, error) {
	var count int
	err := DB.Model(&Subscriber{}).Where("verify_state = ? and subscribe_state = ?", true, true).Count(&count).Error
	return count, err
}

func GetSubscriberByEmail(mail string) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.Find(&subscriber, "email = ?", mail).Error
	return &subscriber, err
}

func GetSubscriberBySignature(key string) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.Find(&subscriber, "signature = ?", key).Error
	return &subscriber, err
}

func GetSubscriberById(id uint) (*Subscriber, error) {
	var subscriber Subscriber
	err := DB.First(&subscriber, id).Error
	return &subscriber, err
}
