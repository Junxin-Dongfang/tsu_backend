package common

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimeToTimestamp 转换 Go time.Time 到 protobuf Timestamp
func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// TimestampToTime 转换 protobuf Timestamp 到 Go time.Time
func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// TimePointerToTimestamp 转换 *time.Time 到 protobuf Timestamp
func TimePointerToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// TimestampToTimePointer 转换 protobuf Timestamp 到 *time.Time
func TimestampToTimePointer(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}