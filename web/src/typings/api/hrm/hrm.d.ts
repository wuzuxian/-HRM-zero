declare namespace Api {
  /**
   * namespace Hrm
   *
   * backend api module: "Hrm"
   */
  namespace Hrm {
    /**
     * org type
     *
     * - "1": 公司
     * - "2": 部门
     */
    type OrgType = '1' | '2';

    /**
     * org status
     *
     * - "1": 开启
     * - "0": 关闭
     */
    type OrgStatus = '0' | '1';

    /** hrm org (组织) tree node */
    type Org = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record create time */
      createTime: string;
      /** record updater */
      updateBy: string;
      /** record update time */
      updateTime: string;
      /** parent org uuid, empty for roots */
      parentUuid: string;
      /** org name */
      orgName: string;
      /** org type: company/department */
      orgType: OrgType;
      /** org leader */
      leader: string;
      /** display order */
      orderNum: number;
      /** org status: open/closed (组织状态 1=开启 0=关闭) */
      status: OrgStatus | null;
      /** remark */
      remark: string;
      /** child orgs */
      children?: Org[];
    };

    /** org search params */
    type OrgSearchParams = CommonType.RecordNullable<Pick<Api.Hrm.Org, 'orgName' | 'orgType' | 'status'>>;

    /** org tree */
    type OrgTree = {
      records: Org[];
    };

    /** add org request */
    type AddOrgRequest = {
      /** parent org uuid, empty for root */
      parentUuid?: string;
      /** org name */
      orgName: string;
      /** org type */
      orgType: OrgType | null;
      /** org leader */
      leader: string;
      /** display order */
      orderNum: number;
      /** remark */
      remark: string;
      /** org status */
      status: OrgStatus | null;
    };

    /** add child org request */
    type AddOrgChildRequest = {
      /** parent org uuid */
      parentUuid: string;
      /** org name */
      orgName: string;
      /** org type */
      orgType: OrgType | null;
      /** org leader */
      leader: string;
      /** display order */
      orderNum: number;
      /** remark */
      remark: string;
      /** org status */
      status: OrgStatus | null;
    };

    /** edit org request */
    type EditOrgRequest = {
      uuid: string;
      /** org name */
      orgName: string;
      /** org type */
      orgType: OrgType | null;
      /** org leader */
      leader: string;
      /** display order */
      orderNum: number;
      /** remark */
      remark: string;
      /** org status */
      status: OrgStatus | null;
    };

    /** import org request: xlsx content encoded as base64 */
    type ImportOrgRequest = {
      fileName: string;
      content: string;
    };

    /** import org response */
    type ImportOrgResponse = {
      successCount: number;
    };

    /** export org response: xlsx content encoded as base64 */
    type ExportOrgResponse = {
      fileName: string;
      content: string;
    };

    /**
     * org chart node (组织架构图节点)
     *
     * - memberCount: 本级人数 (direct employees of the org)
     * - totalMemberCount: 汇总人数 (memberCount + all descendants)
     */
    type OrgChartNode = {
      uuid: string;
      parentUuid: string;
      orgName: string;
      orgType: OrgType;
      leader: string;
      orderNum: number;
      status: OrgStatus | null;
      memberCount: number;
      totalMemberCount: number;
      children?: OrgChartNode[];
    };

    /** org chart tree */
    type OrgChartTree = {
      records: OrgChartNode[];
    };

    /**
     * shift/leave-type status (假勤设置)
     *
     * - "1": 启用
     * - "0": 停用
     */
    type AttendanceStatus = '0' | '1';

    /** hrm shift (班次), 上下班时间支持跨天 */
    type Shift = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** shift name */
      shiftName: string;
      /** start time HH:mm */
      startTime: string;
      /** end time HH:mm, earlier than startTime means cross-day (22:00-次日06:00) */
      endTime: string;
      /** whether the shift crosses midnight: 1 yes 0 no */
      crossDay: number;
      /** rest period start HH:mm, empty when no rest */
      restStart: string;
      /** rest period end HH:mm */
      restEnd: string;
      /** standard working hours = span minus rest, in hours */
      standardHours: number;
      /** flexible minutes, 0 = not flexible */
      flexibleMinutes: number;
      /** status: enabled/disabled */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** shift search params */
    type ShiftSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.Shift, 'shiftName' | 'status'> & Common.CommonSearchParams
    >;

    /** shift paginated list */
    type ShiftList = Common.PaginatingQueryRecord<Shift>;

    /** add shift request */
    type AddShiftRequest = {
      /** shift name */
      shiftName: string;
      /** start time HH:mm */
      startTime: string;
      /** end time HH:mm */
      endTime: string;
      /** rest period start HH:mm */
      restStart: string;
      /** rest period end HH:mm */
      restEnd: string;
      /** flexible minutes */
      flexibleMinutes: number;
      /** status */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
    };

    /** edit shift request */
    type EditShiftRequest = {
      uuid: string;
      shiftName: string;
      startTime: string;
      endTime: string;
      restStart: string;
      restEnd: string;
      flexibleMinutes: number;
      status: AttendanceStatus | null;
      remark: string;
    };

    /** hrm leave type (假期类型) */
    type LeaveType = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** type code: annual_leave/comp_leave/sick_leave/other_leave… */
      typeCode: string;
      /** type name */
      typeName: string;
      /** numeric type value referenced by leave bills */
      typeValue: number;
      /** whether balance management is enabled: 1 yes 0 no */
      balanceEnabled: number;
      /** whether attachment is required: 1 yes 0 no */
      attachmentRequired: number;
      /** default days */
      defaultDays: number;
      /** display order */
      orderNum: number;
      /** status: enabled/disabled */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** leave type search params */
    type LeaveTypeSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.LeaveType, 'typeName' | 'status'> & Common.CommonSearchParams
    >;

    /** leave type paginated list */
    type LeaveTypeList = Common.PaginatingQueryRecord<LeaveType>;

    /** add leave type request */
    type AddLeaveTypeRequest = {
      typeCode: string;
      typeName: string;
      typeValue: number;
      balanceEnabled: number;
      attachmentRequired: number;
      defaultDays: number;
      orderNum: number;
      status: AttendanceStatus | null;
      remark: string;
    };

    /** edit leave type request */
    type EditLeaveTypeRequest = {
      uuid: string;
      typeCode: string;
      typeName: string;
      typeValue: number;
      balanceEnabled: number;
      attachmentRequired: number;
      defaultDays: number;
      orderNum: number;
      status: AttendanceStatus | null;
      remark: string;
    };

    /**
     * HRM bill status (单据状态, shared approval state machine, GLOSSARY)
     *
     * - "unsubmitted": 未提交
     * - "approving": 审批中
     * - "approved": 审批通过
     * - "rejected": 审批不通过
     * - "revoked": 已撤销
     */
    type BillStatus = 'unsubmitted' | 'approving' | 'approved' | 'rejected' | 'revoked';

    /**
     * transfer change type (异动类型)
     *
     * - "post": 调岗
     * - "org": 调部门
     * - "salary": 调薪
     * - "company": 调公司
     * - "other": 其他
     */
    type TransferChangeType = 'post' | 'org' | 'salary' | 'company' | 'other';

    /** hrm transfer bill (人事调动单 HR203) */
    type HrmTransfer = {
      /** record uuid */
      uuid: string;
      /** record creator uuid */
      createBy: string;
      /** record updater uuid */
      updateBy: string;
      /** 单据编号 HR203-... */
      billNo: string;
      /** 单据状态 */
      status: BillStatus | null;
      /** 员工档案 uuid */
      employeeUuid: string;
      /** 员工工号 */
      employeeNo: string;
      /** 员工姓名 */
      employeeName: string;
      /** 异动类型 */
      changeType: TransferChangeType | null;
      /** 原公司（组织 uuid） */
      originalCompany: string;
      /** 原部门（组织 uuid） */
      originalOrg: string;
      /** 原职位 */
      originalPosition: string;
      /** 原岗位 */
      originalPost: string;
      /** 新公司（组织 uuid） */
      newCompany: string;
      /** 新部门（组织 uuid） */
      newOrg: string;
      /** 新职位 */
      newPosition: string;
      /** 新岗位 */
      newPost: string;
      /** 调整后薪资（调薪时填写） */
      newSalary: number;
      /** 调动原因 */
      transferReason: string;
      /** 是否立即生效 1=是 0=否 */
      effectiveImmediately: number;
      /** 生效日期 YYYY-MM-DD */
      effectiveDate: string;
      /** 生效回写是否已完成 1=是 0=否 */
      applied: number;
      /** 申请人 uuid */
      creatorUuid: string;
      /** 申请人姓名 */
      creatorName: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** transfer search params */
    type TransferSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.HrmTransfer, 'billNo' | 'status' | 'employeeNo'> & Common.CommonSearchParams
    >;

    /** transfer paginated list */
    type TransferList = Common.PaginatingQueryRecord<HrmTransfer>;

    /** add transfer request */
    type AddTransferRequest = {
      employeeUuid: string;
      employeeNo: string;
      employeeName: string;
      changeType: TransferChangeType | null;
      originalCompany: string;
      originalOrg: string;
      originalPosition: string;
      originalPost: string;
      newCompany: string;
      newOrg: string;
      newPosition: string;
      newPost: string;
      newSalary: number;
      transferReason: string;
      effectiveImmediately: number;
      effectiveDate: string;
    };

    /** edit transfer request */
    type EditTransferRequest = AddTransferRequest & {
      uuid: string;
    };

    /** shared bill lifecycle action request (提交/审批通过/驳回/撤销) */
    type BillActionRequest = {
      /** bill type code, e.g. "203" for 人事调动 */
      billType: string;
      /** bill uuid */
      billUuid: string;
      /** 审批意见 */
      opinion?: string;
    };

    /** shared approval record (审批记录) */
    type ApprovalRecord = {
      billNo: string;
      /** action: submit/approve/reject/revoke */
      action: string;
      fromStatus: string;
      toStatus: string;
      operatorName: string;
      opinion: string;
      createTime: string;
    };

    /**
     * training session status (培训班次状态)
     *
     * - "1": 未开始
     * - "2": 进行中
     * - "3": 已结束
     */
    type TrainingSessionStatus = '1' | '2' | '3';

    /**
     * training certificate status (培训证书状态)
     *
     * - "0": 待发
     * - "1": 已发
     * - "2": 已撤销
     */
    type TrainingCertificateStatus = '0' | '1' | '2';

    /**
     * training certificate source (证书来源)
     *
     * - "1": 班次结业
     * - "2": 外部导入
     */
    type TrainingCertificateSource = '1' | '2';

    /** my training session (我的班次, read-only view of hrm_training_session) */
    type MyTrainingSession = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 班次编号 PX-年-流水 */
      sessionNo: string;
      /** 班次名称 */
      sessionName: string;
      /** 讲师 */
      lecturer: string;
      /** 地点 */
      location: string;
      /** 学时 */
      hours: number;
      /** 开始时间 */
      startTime: string;
      /** 结束时间 */
      endTime: string;
      /** 状态 */
      status: TrainingSessionStatus | null;
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** my training session search params */
    type MyTrainingSessionSearchParams = CommonType.RecordNullable<
      Pick<MyTrainingSession, 'status'> & Common.CommonSearchParams
    >;

    /** my training session paginated list */
    type MyTrainingSessionList = Common.PaginatingQueryRecord<MyTrainingSession>;

    /** my training certificate (我的培训证书, read-only view of hrm_training_certificate) */
    type MyTrainingCertificate = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 证书号 前缀-年-流水（如 JN-2026-000001） */
      certNo: string;
      /** 证书名称 */
      certName: string;
      /** 证书类型 */
      certType: string;
      /** 来源 1=班次结业 2=外部导入 */
      source: TrainingCertificateSource;
      /** 来源班次 uuid */
      sessionUuid: string;
      /** 发证时间, empty for 待发 */
      issueTime: string;
      /** 状态 0=待发 1=已发 2=已撤销 */
      status: TrainingCertificateStatus | null;
      /** 扫描件文件名, empty when absent */
      fileName: string;
      /** whether a scan file can be previewed/downloaded */
      hasFile: boolean;
      createTime: string;
      updateTime: string;
    };

    /** my training certificate search params */
    type MyTrainingCertificateSearchParams = CommonType.RecordNullable<
      Pick<MyTrainingCertificate, 'status'> & Common.CommonSearchParams
    >;

    /** my training certificate paginated list */
    type MyTrainingCertificateList = Common.PaginatingQueryRecord<MyTrainingCertificate>;

    /** certificate scan file response: content encoded as base64 (mirrors org export) */
    type DownloadMyCertificateFileResponse = {
      fileName: string;
      content: string;
    };
  }
}
