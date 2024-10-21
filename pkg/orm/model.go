package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// EnabledAutoMigrate 是否开启自动迁移
// 每次表迁移耗时，提供此全局变量，程序可根据需要是否迁移
var EnabledAutoMigrate bool

// Scaner 所有模型内组合的结构体，必须满足该接口
type Scaner interface {
	Scan(input interface{}) error
}

// Model int id 模型
// sqlite 不支持 default:now()，支持 CURRENT_TIMESTAMP
type Model struct {
	ID        int  `gorm:"primaryKey;" json:"id"`
	CreatedAt Time `gorm:"notNull;default:CURRENT_TIMESTAMP;index;comment:创建时间" json:"created_at"`
	UpdatedAt Time `gorm:"notNull;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// ModelWithStrID string id 模型
type ModelWithStrID struct {
	ID        string `gorm:"primaryKey;" json:"id"`
	CreatedAt Time   `gorm:"notNull;default:CURRENT_TIMESTAMP;index;comment:创建时间" json:"created_at"`
	UpdatedAt Time   `gorm:"notNull;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (d *ModelWithStrID) BeforeCreate(*gorm.DB) error {
	d.CreatedAt = Now()
	d.UpdatedAt = Now()
	return nil
}

func (d *ModelWithStrID) BeforeUpdate(*gorm.DB) error {
	d.UpdatedAt = Now()
	return nil
}

func (d *Model) BeforeCreate(*gorm.DB) error {
	d.CreatedAt = Now()
	d.UpdatedAt = Now()
	return nil
}

func (d *Model) BeforeUpdate(*gorm.DB) error {
	d.UpdatedAt = Now()
	return nil
}

// NewModelWithStrID 新建模型
func NewModelWithStrID(id string) ModelWithStrID {
	return ModelWithStrID{ID: id, CreatedAt: Now(), UpdatedAt: Now()}
}

// DeletedModel 删除模型
type DeletedModel struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (d *DeletedModel) BeforeCreate(*gorm.DB) error {
	d.CreatedAt = Now()
	d.UpdatedAt = Now()
	return nil
}

func (d *DeletedModel) BeforeUpdate(*gorm.DB) error {
	d.UpdatedAt = Now()
	return nil
}

type DeletedAt = gorm.DeletedAt

type Time struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Time) UnmarshalJSON(b []byte) error {
	date, err := time.ParseInLocation(time.DateTime, strings.Trim(string(b), `"`), time.Local)
	if err == nil {
		t.Time = date
		return nil
	}
	return json.Unmarshal(b, &t.Time)
}

func Now() Time {
	return Time{time.Now()}
}

// Scan implements scaner
func (t *Time) Scan(input interface{}) error {
	var date time.Time
	switch value := input.(type) {
	case time.Time:
		date = value.In(time.FixedZone("CST", 8*60*60))
	case string:
		layout := "2006-01-02 15:04:05"
		aidx := strings.LastIndex(value, ".")
		bidx := strings.LastIndex(value, "+")
		var rear string
		if bidx != -1 {
			rear = "-07:00"
		}
		if aidx != -1 {
			if bidx == -1 {
				bidx = len(value)
			}
			rear = "." + strings.Repeat("9", bidx-aidx) + rear
		}
		layout += rear
		d, err := time.Parse(layout, value)
		if err != nil {
			return fmt.Errorf("pkg: can not convert %v to timestamptz layout[%s]", input, layout)
		}
		date = d.In(time.FixedZone("CST", 8*60*60))
	default:
		return fmt.Errorf("pkg: can not convert %v to timestamptz", input)
	}
	*t = Time{Time: date}
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(time.DateTime) + `"`), nil
}

func (t Time) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// Tabler 模型需要用指针接收器实现接口
type Tabler interface {
	TableName() string
}

// JsonUnmarshal 将 input 反序列化到 obj 上
func JsonUnmarshal(input, obj any) error {
	if v, ok := input.([]byte); ok {
		return json.Unmarshal(v, obj)
	}
	if v, ok := input.(string); ok {
		return json.Unmarshal([]byte(v), obj)
	}
	return nil
}
