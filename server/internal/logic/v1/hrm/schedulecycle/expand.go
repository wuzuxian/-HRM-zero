package schedulecycle

import "time"

// ExpandCycle 将排班周期展开为逐日班次映射（HRM-15 排班表消费契约, 纯函数无 IO）.
//
// 签名: ExpandCycle(cycleShiftIds []int64, startDate, endDate time.Time) map[string]int64
//
// 参数:
//   - cycleShiftIds: 有序班次 id 序列（周期定义, 下标 0 对应起始日）。空序列返回空 map（不报错,
//     周期合法性由 Add/Edit 校验保证）。
//   - startDate: 展开起始日（含）。时间部分被截断为当天 00:00 本地时间, 只用 Y-M-D。
//   - endDate: 展开结束日（含）。若早于 startDate 返回空 map（不报错）。
//
// 返回: key = "YYYY-MM-DD"（dateKey, startDate 按本地时区格式化）,
// value = 当日班次 uuid（cycleShiftIds 中元素原样返回, 语义由调用方约定为 hrm_shift.uuid）。
// 展开规则: 相位锚定本次调用的 startDate（起始日 = 序列第 0 项, 逐日 +1 取模序列长度）,
// 到 endDate 截断。跨月不重置相位; 若 HRM-15 逐月多次调用, 每月相位都从该月锚定日起算
// （想保持绝对相位连续时, 调用方传同一个历史锚点日作为 startDate 即可, 区间外不产出）。
//
// 边界行为（expand_test.go 锁定）:
//   - 跨月: 展开 [2026-01-28, 2026-02-03] 正常逐日产出（月界不重置相位）。
//   - 整除: 区间天数是序列长度的整数倍时, 序列完整重复 N 次。
//   - 非整除截断: 区间在序列中途结束, 尾日相位 = (dayOffset % len)。
//   - startDate == endDate: 恰好产出 1 天。
func ExpandCycle(cycleShiftIds []int64, startDate, endDate time.Time) map[string]int64 {
	out := map[string]int64{}
	if len(cycleShiftIds) == 0 {
		return out
	}
	start := truncateDate(startDate)
	end := truncateDate(endDate)
	if end.Before(start) {
		return out
	}
	day := 24 * time.Hour
	for cur, i := start, 0; !cur.After(end); cur, i = cur.Add(day), i+1 {
		out[cur.Format("2006-01-02")] = cycleShiftIds[i%len(cycleShiftIds)]
	}
	return out
}

// truncateDate 截断到当天 00:00 本地时间（只保留 Y-M-D）.
func truncateDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
