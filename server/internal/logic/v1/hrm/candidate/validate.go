package candidate

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 候选人状态流转（demo-inventory §5 候选人中心）：新线索→已投递→面试中→已发Offer→录用通过。
// 不通过（rejected）可从 除 hired 外 的任意阶段进入（演示站宽松口径：面试未过等），
// hired 为终态，不可再变更。
const (
	CandidateStatusNew          = "new"          // 新线索
	CandidateStatusDelivered    = "delivered"    // 已投递
	CandidateStatusInterviewing = "interviewing" // 面试中
	CandidateStatusOffer        = "offer"        // 已发Offer
	CandidateStatusHired        = "hired"        // 录用通过
	CandidateStatusRejected     = "rejected"     // 不通过
)

var candidateStatuses = map[string]struct{}{
	CandidateStatusNew:          {},
	CandidateStatusDelivered:    {},
	CandidateStatusInterviewing: {},
	CandidateStatusOffer:        {},
	CandidateStatusHired:        {},
	CandidateStatusRejected:     {},
}

// SourceChannel 来源渠道（demo-inventory §5 筛选项）.
const (
	SourceChannelRecommend  = "recommend"  // 内部推荐
	SourceChannelWebsite    = "website"    // 招聘网站
	SourceChannelCampus     = "campus"     // 校园招聘
	SourceChannelHeadhunter = "headhunter" // 猎头
	SourceChannelOther      = "other"      // 其他
)

var sourceChannels = map[string]struct{}{
	SourceChannelRecommend:  {},
	SourceChannelWebsite:    {},
	SourceChannelCampus:     {},
	SourceChannelHeadhunter: {},
	SourceChannelOther:      {},
}

// Education 最高学历（demo-inventory §5 列）.
const (
	EducationDoctor    = "doctor"    // 博士
	EducationMaster    = "master"    // 硕士
	EducationBachelor  = "bachelor"  // 本科
	EducationAssociate = "associate" // 大专
	EducationSenior    = "senior"    // 高中
	EducationOther     = "other"     // 其他
)

var educations = map[string]struct{}{
	EducationDoctor:    {},
	EducationMaster:    {},
	EducationBachelor:  {},
	EducationAssociate: {},
	EducationSenior:    {},
	EducationOther:     {},
}

// forwardOrder is the linear 状态流转 chain: index = stage number. A transition
// is legal iff target == current+1 (不可跳级), no backward moves (不可回退),
// except rejected which any non-hired stage may enter (演示站宽松口径).
var forwardOrder = []string{
	CandidateStatusNew,
	CandidateStatusDelivered,
	CandidateStatusInterviewing,
	CandidateStatusOffer,
	CandidateStatusHired,
}

func statusStage(s string) int {
	for i, v := range forwardOrder {
		if v == s {
			return i
		}
	}
	return -1
}

// validStatusTransition reports whether from → to is a legal 候选人状态流转:
// forward exactly one stage, or to rejected from any stage except hired/rejected.
func validStatusTransition(from, to string) bool {
	if from == to {
		return false
	}
	if from == CandidateStatusHired {
		return false // 终态
	}
	if to == CandidateStatusRejected {
		return from != CandidateStatusRejected
	}
	fi, ti := statusStage(from), statusStage(to)
	if fi < 0 || ti < 0 {
		return false
	}
	return ti == fi+1
}

// validateCandidate checks user-supplied candidate fields, mapping failures to
// the 1044x errcode range.
func validateCandidate(name, sourceChannel, education string) error {
	if name == "" {
		return status.ErrorMessage(errcodes.CandidateInvalidCode, "姓名必填")
	}
	if sourceChannel != "" {
		if _, ok := sourceChannels[sourceChannel]; !ok {
			return status.ErrorMessage(errcodes.CandidateInvalidCode,
				"来源渠道必须是 内部推荐/招聘网站/校园招聘/猎头/其他")
		}
	}
	if education != "" {
		if _, ok := educations[education]; !ok {
			return status.ErrorMessage(errcodes.CandidateInvalidCode,
				"最高学历必须是 博士/硕士/本科/大专/高中/其他")
		}
	}
	return nil
}

func validCandidateStatus(s string) bool {
	_, ok := candidateStatuses[s]
	return ok
}
