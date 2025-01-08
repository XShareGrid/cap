package utils

import cap "github.com/XShareGrid/cap/table/proto/go"

// DoMemoryPaging 内存数据分页处理
func DoMemoryPaging(results []interface{}, pageParam *cap.PageParam) ([]interface{}, *cap.PageInfo) {
	pageInfo := &cap.PageInfo{}
	pageInfo.TotalResults = int32(len(results))
	pageInfo.CurrentPage = pageParam.Page
	pageInfo.PageSize = pageParam.PageSize
	pageInfo.TotalPages = 1
	if pageParam.PageSize > 0 {
		pageInfo.TotalPages = int32(len(results)) / pageParam.PageSize
		if int32(len(results))%pageParam.PageSize > 0 {
			pageInfo.TotalPages++
		}
		startIdx := int(pageParam.PageSize) * int(pageParam.Page)
		endIdx := startIdx + int(pageParam.PageSize)
		if startIdx > len(results) {
			results = []interface{}{}
		} else {
			if endIdx > len(results) {
				endIdx = len(results)
			}
			results = results[startIdx:endIdx]
		}
	}
	return results, pageInfo
}
