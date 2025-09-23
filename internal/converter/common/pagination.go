package common

import (
	"tsu-self/internal/repository/query"
	"tsu-self/internal/rpc/generated/common"
)

// PaginationToRPC 转换查询分页参数到 RPC
func PaginationToRPC(pagination *query.Pagination) *common.Pagination {
	if pagination == nil {
		return nil
	}

	return &common.Pagination{
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}
}

// PaginationFromRPC 转换 RPC 分页参数到查询参数
func PaginationFromRPC(rpcPagination *common.Pagination) *query.Pagination {
	if rpcPagination == nil {
		return &query.Pagination{
			Page:     1,
			PageSize: 20,
		}
	}

	pagination := &query.Pagination{
		Page:     int(rpcPagination.Page),
		PageSize: int(rpcPagination.PageSize),
	}
	pagination.Validate()
	return pagination
}

// PaginationResultToRPC 转换分页结果到 RPC
func PaginationResultToRPC(result *query.PaginationResult) *common.Pagination {
	if result == nil {
		return nil
	}

	return &common.Pagination{
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		Total:      result.Total,
		TotalPages: int32(result.TotalPages),
	}
}