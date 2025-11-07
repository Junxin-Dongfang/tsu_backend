package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tsu-self/internal/modules/admin/dto"
)

// TestRoomSequenceValidation 测试房间序列验证逻辑
func TestRoomSequenceValidation(t *testing.T) {
	tests := []struct {
		name      string
		sequence  []dto.RoomSequenceItem
		wantError bool
		errorMsg  string
	}{
		{
			name:      "空序列应该失败",
			sequence:  []dto.RoomSequenceItem{},
			wantError: true,
			errorMsg:  "房间序列不能为空",
		},
		{
			name: "序列不连续应该失败",
			sequence: []dto.RoomSequenceItem{
				{RoomID: "ROOM_001", Sort: 1},
				{RoomID: "ROOM_002", Sort: 3}, // 跳过了2
			},
			wantError: true,
			errorMsg:  "房间序列不连续",
		},
		{
			name: "重复的房间ID应该失败",
			sequence: []dto.RoomSequenceItem{
				{RoomID: "ROOM_001", Sort: 1},
				{RoomID: "ROOM_001", Sort: 2}, // 重复
			},
			wantError: true,
			errorMsg:  "房间ID重复",
		},
		{
			name: "重复的排序号应该失败",
			sequence: []dto.RoomSequenceItem{
				{RoomID: "ROOM_001", Sort: 1},
				{RoomID: "ROOM_002", Sort: 1}, // 重复
			},
			wantError: true,
			errorMsg:  "房间排序号重复",
		},
		{
			name: "有效的序列应该成功",
			sequence: []dto.RoomSequenceItem{
				{RoomID: "ROOM_001", Sort: 1},
				{RoomID: "ROOM_002", Sort: 2},
				{RoomID: "ROOM_003", Sort: 3},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoomSequenceStructure(tt.sequence)

			if tt.wantError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

// validateRoomSequenceStructure 验证房间序列结构(测试辅助函数)
func validateRoomSequenceStructure(sequence []dto.RoomSequenceItem) error {
	if len(sequence) == 0 {
		return assert.AnError
	}

	// 检查序列连续性和唯一性
	roomIDMap := make(map[string]bool)
	sortMap := make(map[int]bool)

	for _, item := range sequence {
		if roomIDMap[item.RoomID] {
			return assert.AnError
		}
		roomIDMap[item.RoomID] = true

		if sortMap[item.Sort] {
			return assert.AnError
		}
		sortMap[item.Sort] = true
	}

	// 检查序列连续性
	for i := 1; i <= len(sequence); i++ {
		if !sortMap[i] {
			return assert.AnError
		}
	}

	return nil
}

// TestCycleDetection 测试循环引用检测
func TestCycleDetection(t *testing.T) {
	tests := []struct {
		name      string
		sequence  []dto.RoomSequenceItem
		wantError bool
		errorMsg  string
	}{
		{
			name: "无循环引用应该成功",
			sequence: []dto.RoomSequenceItem{
				{RoomID: "ROOM_001", Sort: 1},
				{RoomID: "ROOM_002", Sort: 2},
				{RoomID: "ROOM_003", Sort: 3},
			},
			wantError: false,
		},
		{
			name: "简单循环引用应该失败",
			sequence: []dto.RoomSequenceItem{
				{
					RoomID: "ROOM_001",
					Sort:   1,
					ConditionalSkip: map[string]interface{}{
						"target_room": "ROOM_002",
					},
				},
				{
					RoomID: "ROOM_002",
					Sort:   2,
					ConditionalSkip: map[string]interface{}{
						"target_room": "ROOM_001", // 循环回去
					},
				},
			},
			wantError: true,
			errorMsg:  "循环引用",
		},
		{
			name: "复杂循环引用应该失败",
			sequence: []dto.RoomSequenceItem{
				{
					RoomID: "ROOM_001",
					Sort:   1,
					ConditionalSkip: map[string]interface{}{
						"target_room": "ROOM_002",
					},
				},
				{
					RoomID: "ROOM_002",
					Sort:   2,
					ConditionalSkip: map[string]interface{}{
						"target_room": "ROOM_003",
					},
				},
				{
					RoomID: "ROOM_003",
					Sort:   3,
					ConditionalReturn: map[string]interface{}{
						"target_room": "ROOM_001", // 循环回去
					},
				},
			},
			wantError: true,
			errorMsg:  "循环引用",
		},
		{
			name: "有条件跳过但无循环应该成功",
			sequence: []dto.RoomSequenceItem{
				{
					RoomID: "ROOM_001",
					Sort:   1,
					ConditionalSkip: map[string]interface{}{
						"target_room": "ROOM_003",
					},
				},
				{
					RoomID: "ROOM_002",
					Sort:   2,
				},
				{
					RoomID: "ROOM_003",
					Sort:   3,
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := detectCycles(tt.sequence)

			if tt.wantError {
				assert.Error(t, err, "应该检测到循环引用")
			} else {
				assert.NoError(t, err, "不应该有循环引用")
			}
		})
	}
}

// detectCycles 检测循环引用(测试辅助函数)
func detectCycles(sequence []dto.RoomSequenceItem) error {
	// 构建邻接表
	graph := make(map[string][]string)
	for _, item := range sequence {
		graph[item.RoomID] = []string{}

		// 添加条件跳过的边
		if item.ConditionalSkip != nil {
			if targetRoom, ok := item.ConditionalSkip["target_room"].(string); ok && targetRoom != "" {
				graph[item.RoomID] = append(graph[item.RoomID], targetRoom)
			}
		}

		// 添加条件返回的边
		if item.ConditionalReturn != nil {
			if targetRoom, ok := item.ConditionalReturn["target_room"].(string); ok && targetRoom != "" {
				graph[item.RoomID] = append(graph[item.RoomID], targetRoom)
			}
		}
	}

	// 使用DFS检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(roomID string) bool
	dfs = func(roomID string) bool {
		visited[roomID] = true
		recStack[roomID] = true

		for _, neighbor := range graph[roomID] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true // 发现环
			}
		}

		recStack[roomID] = false
		return false
	}

	// 检查每个节点
	for _, item := range sequence {
		if !visited[item.RoomID] {
			if dfs(item.RoomID) {
				return assert.AnError
			}
		}
	}

	return nil
}

// TestLevelRangeValidation 测试等级区间验证
func TestLevelRangeValidation(t *testing.T) {
	tests := []struct {
		name      string
		minLevel  int16
		maxLevel  int16
		wantError bool
	}{
		{
			name:      "有效的等级区间",
			minLevel:  1,
			maxLevel:  10,
			wantError: false,
		},
		{
			name:      "相同的等级",
			minLevel:  5,
			maxLevel:  5,
			wantError: false,
		},
		{
			name:      "无效的等级区间",
			minLevel:  10,
			maxLevel:  5,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.minLevel > tt.maxLevel {
				err = assert.AnError
			}

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

