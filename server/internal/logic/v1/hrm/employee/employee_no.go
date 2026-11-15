package employee

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 工号规则 (spec user story 10): 为空时按「最大工号+1」自动生成, 8 位补零。
// 纯函数 seam: NextEmployeeNo / maxNumericEmployeeNo 不依赖 DB, 有独立单测。
const (
	// employeeNoWidth 工号固定 8 位, 左侧补零
	employeeNoWidth = 8
	// employeeNoCeiling 8 位补零工号的最大值 (00000001..99999999)
	employeeNoCeiling = 99999999
)

var (
	// ErrEmployeeNoExhausted 当前最大工号已到 99999999, 无法再生成
	ErrEmployeeNoExhausted = errors.New("工号流水已超出 8 位上限 (99999999)")
	// ErrEmployeeNoInvalid 传入的最大工号不是纯数字, 无法递增
	ErrEmployeeNoInvalid = errors.New("最大工号必须是纯数字")
)

// NextEmployeeNo 根据当前最大工号计算下一个工号: 最大工号+1, 8 位补零。
// maxNo 为空 (空表/没有数字工号) 时返回 "00000001"。
func NextEmployeeNo(maxNo string) (string, error) {
	trimmed := strings.TrimSpace(maxNo)
	if trimmed == "" {
		return fmt.Sprintf("%0*d", employeeNoWidth, 1), nil
	}
	n, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || n < 0 {
		return "", ErrEmployeeNoInvalid
	}
	if n >= employeeNoCeiling {
		return "", ErrEmployeeNoExhausted
	}
	return fmt.Sprintf("%0*d", employeeNoWidth, n+1), nil
}

// maxNumericEmployeeNo 在既有工号集合中找到数值最大的工号 (跳过非数字工号, 如
// 导入的历史编号 "HR-001"); 集合中没有数字工号时返回 ""。
func maxNumericEmployeeNo(employeeNos []string) string {
	best := int64(-1)
	bestRaw := ""
	for _, no := range employeeNos {
		trimmed := strings.TrimSpace(no)
		if trimmed == "" {
			continue
		}
		if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil && n > best {
			best = n
			bestRaw = trimmed
		}
	}
	return bestRaw
}
