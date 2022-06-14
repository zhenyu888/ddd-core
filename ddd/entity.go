package ddd

import "time"

type Entity interface {
	Identifier() int64
}

type MixModel struct {
	Id         int64     `gorm:"primaryKey"`
	CreateTime time.Time `gorm:"autoCreateTime"`
	UpdateTime time.Time `gorm:"autoUpdateTime"`
}

func (a *MixModel) AggregateId() int64 {
	return a.Id
}

func (a *MixModel) Identifier() int64 {
	return a.Id
}
