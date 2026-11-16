// Package errcodes — HRM-27 录用审批单（HR208）业务错误码（1065x 段）。
// 开工时 grep 确认 1065x 空闲（1060x 加班 / 1061x 薪酬方案+看板 / 1062x 薪资档案 /
// 1063x 绩效批次 / 1064x 培训签到已占用）；菜单 uuid 段 34xx，迁移 48。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// OfferNotFoundCode 录用审批单不存在
	OfferNotFoundCode = 10650
	// OfferInvalidCode 录用审批单参数非法（候选人/职位/日期必填等）
	OfferInvalidCode = 10651
	// OfferNotEditableCode 当前单据状态不允许修改/删除（仅 未提交/审批不通过/已撤销）
	OfferNotEditableCode = 10652
	// OfferCandidateNotOfferableCode 候选人不在 offer（已发Offer）阶段，不可发起录用
	OfferCandidateNotOfferableCode = 10653
	// OfferPositionNotOpenableCode 招聘职位不在 open（招聘中）状态，不可发起录用
	OfferPositionNotOpenableCode = 10654
)

func init() {
	status.Register(OfferNotFoundCode)
	status.Register(OfferInvalidCode)
	status.Register(OfferNotEditableCode)
	status.Register(OfferCandidateNotOfferableCode)
	status.Register(OfferPositionNotOpenableCode)
}
