package web

import (
	"fmt"
	"strings"
)

// ScrollPager 滚动翻页
type ScrollPager[T any] struct {
	Data []T    `json:"data"`
	Next string `json:"next"`
}

// PageOutput 分页数据
type PageOutput struct {
	Total int64 `json:"total"`
	Items any   `json:"items"`
}

// PagerFilter 分页过滤
type PagerFilter struct {
	Page         int      `form:"page"`
	Size         int      `form:"size"`
	Sort         string   `form:"sort"`
	SortSafelist []string `json:"-"`
}

// MustSortColumn 忽略安全问题
func (f PagerFilter) MustSortColumn() string {
	return strings.TrimLeft(f.Sort, "-")
}

// SortColumn 通过对 SortColumn 设置值，仅对允许的值做排序处理
func (f PagerFilter) SortColumn() (string, error) {
	for _, v := range f.SortSafelist {
		if f.Sort == v {
			return strings.TrimLeft(f.Sort, "-"), nil
		}
	}
	return "", fmt.Errorf("%s 不支持排序", f.Sort)
}

// SortDirection 如果 sort 携带负号返回倒序，否则返回正序
func (f PagerFilter) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// Offset 计算偏离数值
func (f PagerFilter) Offset() int {
	if f.Page < 1 {
		f.Page = 1
	}
	return (f.Page - 1) * f.Size
}

// Limit 每页 10~100 区间
func (f PagerFilter) Limit() int {
	if f.Size <= 1 {
		return 10
	}
	if f.Size > 10000 {
		return 10000
	}
	return f.Size
}

func Limit(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func Offset(page, size int) int {
	if page < 1 {
		return 1
	}
	return (page - 1) * size
}
