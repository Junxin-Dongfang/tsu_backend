package interfaces

import "errors"

var (
	// ErrTeamDungeonProgressNotFound 进度不存在
	ErrTeamDungeonProgressNotFound = errors.New("team dungeon progress not found")
	// ErrTeamDungeonRecordNotFound 记录不存在
	ErrTeamDungeonRecordNotFound = errors.New("team dungeon record not found")
)
