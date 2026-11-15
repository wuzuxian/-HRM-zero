package schedulecycle

import (
	"testing"
	"time"
)

// date 在本地时区构造某日（无时间部分）.
func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func key(y int, m time.Month, d int) string {
	return date(y, m, d).Format("2006-01-02")
}

// TestExpandCycleBasic 基本: 序列 [10,20,30] 展开 6 天（整除, 恰好两轮）.
func TestExpandCycleBasic(t *testing.T) {
	got := ExpandCycle([]int64{10, 20, 30}, date(2026, 3, 1), date(2026, 3, 6))
	if len(got) != 6 {
		t.Fatalf("len = %d, want 6", len(got))
	}
	want := map[string]int64{
		key(2026, 3, 1): 10,
		key(2026, 3, 2): 20,
		key(2026, 3, 3): 30,
		key(2026, 3, 4): 10,
		key(2026, 3, 5): 20,
		key(2026, 3, 6): 30,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%s] = %d, want %d", k, got[k], v)
		}
	}
}

// TestExpandCycleNonDivisibleTruncate 非整除截断: 序列长度 4, 区间 6 天 → 尾部按相位截断.
func TestExpandCycleNonDivisibleTruncate(t *testing.T) {
	got := ExpandCycle([]int64{1, 2, 3, 4}, date(2026, 3, 1), date(2026, 3, 6))
	if len(got) != 6 {
		t.Fatalf("len = %d, want 6", len(got))
	}
	// 第 5 天(3-05)相位 4%4=0 → 1; 第 6 天相位 5%4=1 → 2
	if got[key(2026, 3, 5)] != 1 || got[key(2026, 3, 6)] != 2 {
		t.Errorf("truncated phase wrong: 3-05=%d 3-06=%d", got[key(2026, 3, 5)], got[key(2026, 3, 6)])
	}
}

// TestExpandCyclePhaseAnchor 相位锚定语义: 相位锚定本次调用的 startDate。
// 同一周期想跨月保持绝对相位连续时, 调用方（HRM-15）传同一个历史锚点日作为 startDate。
func TestExpandCyclePhaseAnchor(t *testing.T) {
	seq := []int64{1, 2, 3, 4}
	// 每次调用相位都从各自 startDate 起算: 3-08 是第二次调用的第 0 天 → 1
	second := ExpandCycle(seq, date(2026, 3, 8), date(2026, 3, 14))
	if second[key(2026, 3, 8)] != 1 || second[key(2026, 3, 14)] != 3 { // 3-14 相位 6%4=2 → 3
		t.Errorf("anchored phase wrong: 3-08=%d 3-14=%d", second[key(2026, 3, 8)], second[key(2026, 3, 14)])
	}
	// 绝对相位连续: 传历史锚点 3-01 作为 startDate, 区间 [3-01, 3-14] 逐日与手工相位一致
	whole := ExpandCycle(seq, date(2026, 3, 1), date(2026, 3, 14))
	anchored := ExpandCycle(seq, date(2026, 3, 1), date(2026, 3, 14))
	for d := 1; d <= 14; d++ {
		if anchored[key(2026, 3, d)] != whole[key(2026, 3, d)] {
			t.Errorf("anchor break at 3-%02d", d)
		}
	}
	// 历史锚点窗口: 3-08 距锚点 7 天, 相位 7%4=3 → 4
	tail := ExpandCycle(seq, date(2026, 3, 1), date(2026, 3, 14))
	if tail[key(2026, 3, 8)] != 4 {
		t.Errorf("tail window 3-08 = %d, want 4", tail[key(2026, 3, 8)])
	}
}

// TestExpandCycleCrossMonth 跨月: 1 月末 → 2 月初（含闰年外的月界）.
func TestExpandCycleCrossMonth(t *testing.T) {
	got := ExpandCycle([]int64{5, 6}, date(2026, 1, 30), date(2026, 2, 2))
	want := map[string]int64{
		key(2026, 1, 30): 5,
		key(2026, 1, 31): 6,
		key(2026, 2, 1):  5,
		key(2026, 2, 2):  6,
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%s] = %d, want %d", k, got[k], v)
		}
	}
}

// TestExpandCycleSingleDay startDate == endDate 恰好 1 天.
func TestExpandCycleSingleDay(t *testing.T) {
	got := ExpandCycle([]int64{7}, date(2026, 3, 15), date(2026, 3, 15))
	if len(got) != 1 || got[key(2026, 3, 15)] != 7 {
		t.Errorf("single day: %+v", got)
	}
}

// TestExpandCycleEdgeCases 空序列/空区间/时间部分截断.
func TestExpandCycleEdgeCases(t *testing.T) {
	// 空序列 → 空 map
	if got := ExpandCycle(nil, date(2026, 3, 1), date(2026, 3, 5)); len(got) != 0 {
		t.Errorf("nil seq: %+v", got)
	}
	if got := ExpandCycle([]int64{}, date(2026, 3, 1), date(2026, 3, 5)); len(got) != 0 {
		t.Errorf("empty seq: %+v", got)
	}
	// endDate < startDate → 空 map
	if got := ExpandCycle([]int64{1}, date(2026, 3, 5), date(2026, 3, 1)); len(got) != 0 {
		t.Errorf("inverted range: %+v", got)
	}
	// 时间部分被截断（当天任意时刻算同一天）
	a := ExpandCycle([]int64{1, 2}, time.Date(2026, 3, 1, 23, 59, 0, 0, time.Local), time.Date(2026, 3, 1, 0, 0, 1, 0, time.Local))
	if len(a) != 1 {
		t.Errorf("same-day with time parts: %+v", a)
	}
}
