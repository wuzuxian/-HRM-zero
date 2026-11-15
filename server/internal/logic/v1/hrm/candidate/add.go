package candidate

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增候选人（spec 用户故事 45）：初始状态 新线索（new）；来源渠道缺省 other.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateCandidate(req.Name, req.SourceChannel, req.Education); err != nil {
		return nil, err
	}
	if req.SourceChannel == "" {
		req.SourceChannel = SourceChannelOther
	}

	if err = l.svcCtx.Model.HrmRecruitmentCandidate.InsertV2(l.ctx, nil, &hrm_recruitment_candidate.HrmRecruitmentCandidate{
		Uuid:            uuid.New().String(),
		Name:            req.Name,
		Phone:           req.Phone,
		Email:           req.Email,
		SourceChannel:   req.SourceChannel,
		Education:       req.Education,
		CurrentCompany:  req.CurrentCompany,
		CurrentPosition: req.CurrentPosition,
		ExpectSalary:    req.ExpectSalary,
		Status:          CandidateStatusNew,
		Remark:          req.Remark,
	}); err != nil {
		return nil, err
	}

	l.Infof("candidate: 候选人已创建 name=%s source=%s", req.Name, req.SourceChannel)
	return &types.AddResponse{}, nil
}

// candidateRecord converts a model row into the API record.
func candidateRecord(row *hrm_recruitment_candidate.HrmRecruitmentCandidate) types.HrmCandidate {
	return types.HrmCandidate{
		Uuid:            row.Uuid,
		Name:            row.Name,
		Phone:           row.Phone,
		Email:           row.Email,
		SourceChannel:   row.SourceChannel,
		Education:       row.Education,
		CurrentCompany:  row.CurrentCompany,
		CurrentPosition: row.CurrentPosition,
		ExpectSalary:    row.ExpectSalary,
		Status:          row.Status,
		CreateTime:      row.CreateTime.Format(time.DateTime),
		UpdateTime:      row.UpdateTime.Format(time.DateTime),
	}
}
