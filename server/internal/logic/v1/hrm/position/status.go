package position

// 职位状态（demo-inventory §5 招聘职位：状态(招聘中)）.
const (
	PositionStatusOpen   = "open"   // 招聘中
	PositionStatusFull   = "full"   // 已满
	PositionStatusClosed = "closed" // 已关闭
)

var positionStatuses = map[string]struct{}{
	PositionStatusOpen:   {},
	PositionStatusFull:   {},
	PositionStatusClosed: {},
}

// validPositionStatus reports whether s is a legal 职位状态 value.
func validPositionStatus(s string) bool {
	_, ok := positionStatuses[s]
	return ok
}
