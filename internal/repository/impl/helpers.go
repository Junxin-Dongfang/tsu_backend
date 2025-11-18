package impl

import "time"

// nullTimeNow 返回当前时间的指针（用于 null.Time）
func nullTimeNow() *time.Time {
	now := time.Now()
	return &now
}

