package signin

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type ExportSigninDetail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewExportSigninDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ExportSigninDetail {
	return &ExportSigninDetail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// exportHeader 签到明细导出列（demo-inventory 6：姓名/工号/部门/签到时间/方式/操作人/原因 + 账号定位）。
var exportSigninHeader = []string{"姓名", "登录账号", "工号", "部门", "签到时间", "签到方式", "操作人", "补签原因"}

// ExportSigninDetail 导出签到明细 xlsx（base64，复用 employee/punch 导出的 excelize 惯例）。
// 口径与 GetSigninDetailList 完全一致（名册=应到名单，未签行签到列留空）。
func (l *ExportSigninDetail) ExportSigninDetail(req *types.ExportSigninDetailRequest) (resp *types.ExportSigninDetailResponse, err error) {
	detail, err := NewGetSigninDetailList(l.ctx, l.svcCtx, l.r).GetSigninDetailList(
		&types.GetSigninDetailListRequest{SessionUuid: req.SessionUuid})
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err = f.SetSheetRow(sheet, "A1", &exportSigninHeader); err != nil {
		return nil, err
	}
	for i, rec := range detail.Records {
		line := []string{
			rec.NickName, rec.Username, rec.EmployeeNo, rec.OrgName,
			rec.SignInTime, signinMethodName(rec.SignInMethod), rec.SignInByName, rec.SignInReason,
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err = f.SetSheetRow(sheet, cell, &line); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err = f.Write(&buf); err != nil {
		return nil, err
	}
	return &types.ExportSigninDetailResponse{
		FileName: fmt.Sprintf("signin-detail-%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// signinMethodName 签到方式展示名。
func signinMethodName(v string) string {
	switch v {
	case SignInMethodSelf:
		return "扫码签到"
	case SignInMethodFor:
		return "管理员补签"
	}
	return ""
}
