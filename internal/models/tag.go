package models

type Tag struct {
	Id   string `json:"id" gorm:"primaryKey;not null;uniqueIndex"`
	Name string `json:"name" gorm:"type:varchar(50);not null;uniqueIndex"`
}

func (Tag) TableName() string {
	return "tags"
}
