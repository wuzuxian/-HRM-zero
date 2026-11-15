package employee

import (
	"testing"
)

// TestNextEmployeeNo 工号自动生成纯函数 (spec user story 10): 空/补零/递增/并发序号边界.
func TestNextEmployeeNo(t *testing.T) {
	cases := []struct {
		name    string
		maxNo   string
		want    string
		wantErr error
	}{
		{name: "空表→00000001", maxNo: "", want: "00000001"},
		{name: "空格→00000001", maxNo: "   ", want: "00000001"},
		{name: "补零递增", maxNo: "00000005", want: "00000006"},
		{name: "不足8位补零", maxNo: "17", want: "00000018"},
		{name: "全零后进位", maxNo: "00099999", want: "00100000"},
		{name: "8位上限", maxNo: "99999999", wantErr: ErrEmployeeNoExhausted},
		{name: "非数字非法", maxNo: "HR-001", wantErr: ErrEmployeeNoInvalid},
		{name: "负数非法", maxNo: "-1", wantErr: ErrEmployeeNoInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NextEmployeeNo(c.maxNo)
			if c.wantErr != nil {
				if err != c.wantErr {
					t.Fatalf("NextEmployeeNo(%q) err = %v, want %v", c.maxNo, err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NextEmployeeNo(%q) unexpected err: %v", c.maxNo, err)
			}
			if got != c.want {
				t.Fatalf("NextEmployeeNo(%q) = %q, want %q", c.maxNo, got, c.want)
			}
		})
	}
}

// TestMaxNumericEmployeeNo 最大工号扫描: 跳过非数字历史编号 (如导入的 HR-001).
func TestMaxNumericEmployeeNo(t *testing.T) {
	cases := []struct {
		name string
		nos  []string
		want string
	}{
		{name: "空", nos: nil, want: ""},
		{name: "全部非数字", nos: []string{"HR-001", "A002"}, want: ""},
		{name: "混合", nos: []string{"00000009", "HR-001", "00000012", "00000003"}, want: "00000012"},
		{name: "数值比较而非字典序", nos: []string{"9", "00000012"}, want: "00000012"},
		{name: "空白跳过", nos: []string{"", "  ", "00000007"}, want: "00000007"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := maxNumericEmployeeNo(c.nos); got != c.want {
				t.Fatalf("maxNumericEmployeeNo(%v) = %q, want %q", c.nos, got, c.want)
			}
		})
	}
}
