package sharedmodel

// 数据库中用到的枚举定义
type GenderEnum string

const (
	Male           GenderEnum = "male"
	Female         GenderEnum = "female"
	Other          GenderEnum = "other"
	PreferNotToSay GenderEnum = "prefer_not_to_say"
)
