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

    // ---- HRM 培训计划单 (HRM-28, HR216) ----

    /**
     * training plan type (培训计划 计划类型)
     *
     * - "special": 专项
     * - "new_hire": 新员工培养
     * - "safety": 安全培训
     * - "skill": 技能提升
     */
    type TrainingPlanType = 'special' | 'new_hire' | 'safety' | 'skill';

    /** hrm training plan (培训计划单 HR216, 单据审批底座 + 发布状态) */
    type TrainingPlan = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 单据编号 HR216-... */
      billNo: string;
      /** 单据状态 */
      status: BillStatus | null;
      /** 计划名称 */
      planName: string;
      /** 主题 */
      theme: string;
      /** 年度 */
      planYear: string;
      /** 计划类型 */
      planType: TrainingPlanType | null;
      /** 开始日期 YYYY-MM-DD */
      startDate: string;
      /** 结束日期 YYYY-MM-DD */
      endDate: string;
      /** 发布状态 0=草稿 1=已发布（审批通过即发布） */
      published: number;
      /** 申请人 uuid */
      creatorUuid: string;
      /** 申请人姓名 */
      creatorName: string;
      createTime: string;
      updateTime: string;
    };

    /** training plan search params (published 过滤为 '0'/'1' 字符串) */
    type TrainingPlanSearchParams = CommonType.RecordNullable<
      Pick<TrainingPlan, 'planName' | 'planType' | 'status'> & Common.CommonSearchParams
    > & {
      /** 发布状态 '0'=草稿 '1'=已发布 */
      published?: string | null;
    };

    /** training plan paginated list */
    type TrainingPlanList = Common.PaginatingQueryRecord<TrainingPlan>;

    /** add training plan request */
    type AddTrainingPlanRequest = {
      planName: string;
      theme: string;
      planYear: string;
      planType: TrainingPlanType | null;
      startDate: string;
      endDate: string;
    };

    /** edit training plan request */
    type EditTrainingPlanRequest = AddTrainingPlanRequest & {
      uuid: string;
    };

    // ---- HRM 培训班次 (HRM-28, 管理端 CRUD) ----

    /** hrm training session (培训班次, admin view of hrm_training_session) */
    type TrainingSession = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 班次编号 PX-年-流水（自动生成） */
      sessionNo: string;
      /** 班次名称 */
      sessionName: string;
      /** 关联培训计划 uuid */
      planUuid: string;
      /** 关联培训计划名称（展示快照） */
      planName: string;
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

    /** training session search params */
    type TrainingSessionSearchParams = CommonType.RecordNullable<
      Pick<TrainingSession, 'sessionName' | 'planUuid' | 'status'> & Common.CommonSearchParams
    >;

    /** training session paginated list */
    type TrainingSessionList = Common.PaginatingQueryRecord<TrainingSession>;

    /** add training session request (PX 编号后端自动生成) */
    type AddTrainingSessionRequest = {
      sessionName: string;
      planUuid: string;
      lecturer: string;
      location: string;
      hours: number;
      startTime: string;
      endTime: string;
      remark: string;
    };

    /** edit training session request */
    type EditTrainingSessionRequest = AddTrainingSessionRequest & {
      uuid: string;
    };

    /** session participant row (班次学员) */
    type SessionParticipant = {
      /** participant row uuid */
      uuid: string;
      /** 学员（manage_user.uuid） */
      userUuid: string;
      /** 学员账号 */
      username: string;
      /** 学员姓名 */
      nickName: string;
    };

    /** session participant list response */
    type SessionParticipantList = {
      records: SessionParticipant[];
    };

    /** set session participants request (选人, 全量覆盖) */
    type SetSessionParticipantsRequest = {
      sessionUuid: string;
      userUuids: string[];
    };

    // ---- HRM 员工档案 (HRM-03) ----

    /** 性别 1=男 2=女 */
    type Gender = '1' | '2';

    /** 婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶 */
    type MaritalStatus = '1' | '2' | '3' | '4';

    /** 人员状态 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休 (5/6 = 不在册口径) */
    type PersonStatus = '1' | '2' | '3' | '4' | '5' | '6';

    /** 用户状态 1=已生成系统账号 0=未生成 (由 user_uuid 派生) */
    type UserStatus = '0' | '1';

    /** 教育经历从表行 */
    type EmployeeEducation = {
      uuid?: string;
      school: string;
      major: string;
      education: string;
      startDate: string;
      endDate: string;
      remark: string;
    };

    /** 工作经历从表行 */
    type EmployeeWork = {
      uuid?: string;
      company: string;
      position: string;
      startDate: string;
      endDate: string;
      remark: string;
    };

    /** 家属从表行 */
    type EmployeeFamily = {
      uuid?: string;
      name: string;
      relation: string;
      phone: string;
      workUnit: string;
      remark: string;
    };

    /** 员工档案主表行 (列表不带从表, 详情带) */
    type Employee = {
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
      /** employeeNo */
      employeeNo: string;
      name: string;
      gender: Gender | '';
      birthday: string;
      nation: string;
      politicalStatus: string;
      maritalStatus: MaritalStatus | '';
      nativePlace: string;
      bloodType: string;
      idCard: string;
      height: string;
      weight: string;
      householdAddress: string;
      address: string;
      emergencyContact: string;
      emergencyPhone: string;
      phone: string;
      email: string;
      avatar: string;
      bankName: string;
      bankAccount: string;
      companyUuid: string;
      orgUuid: string;
      position: string;
      post: string;
      personStatus: PersonStatus;
      hireDate: string;
      formalDate: string;
      /** 用户状态: 是否已生成系统账号 (1=已生成 0=未生成) */
      userStatus: UserStatus;
      remark: string;
      educations?: EmployeeEducation[];
      works?: EmployeeWork[];
      families?: EmployeeFamily[];
    };

    /** 员工档案 search params */
    type EmployeeSearchParams = CommonType.RecordNullable<
      Pick<Employee, 'employeeNo' | 'name' | 'orgUuid' | 'personStatus'> & Common.CommonSearchParams
    >;

    /** 员工档案 paginated list */
    type EmployeeList = Common.PaginatingQueryRecord<Employee>;

    /** add employee request; employeeNo 为空时后端按 最大工号+1 自动生成 */
    type AddEmployeeRequest = {
      employeeNo?: string;
      name: string;
      gender?: Gender | '';
      birthday?: string;
      nation?: string;
      politicalStatus?: string;
      maritalStatus?: MaritalStatus | '';
      nativePlace?: string;
      bloodType?: string;
      idCard?: string;
      height?: string;
      weight?: string;
      householdAddress?: string;
      address?: string;
      emergencyContact?: string;
      emergencyPhone?: string;
      phone?: string;
      email?: string;
      avatar?: string;
      bankName?: string;
      bankAccount?: string;
      companyUuid?: string;
      orgUuid?: string;
      position?: string;
      post?: string;
      personStatus?: PersonStatus;
      hireDate?: string;
      formalDate?: string;
      remark?: string;
      educations?: EmployeeEducation[];
      works?: EmployeeWork[];
      families?: EmployeeFamily[];
    };

    /** edit employee request (从表全删全插) */
    type EditEmployeeRequest = AddEmployeeRequest & {
      uuid: string;
    };

    /** employee detail response */
    type EmployeeDetailResponse = {
      employee: Employee;
    };

    /** import employees request (base64 xlsx, mirrors org import) */
    type ImportEmployeeRequest = {
      fileName: string;
      content: string;
    };

    type ImportEmployeeResponse = {
      successCount: number;
    };

    /** export employees response (base64 xlsx, mirrors org export) */
    type ExportEmployeeResponse = {
      fileName: string;
      content: string;
    };

    /** 生成系统账号请求 (HRM-04): 单个/批量; 已生成过账号的员工自动跳过 */
    type GenerateUserRequest = {
      uuids: string[];
    };

    type GenerateUserResponse = {
      /** 本次实际新生成的账号数 */
      generated: number;
      /** 已有账号而跳过的员工数 */
      skipped: number;
    };
    /**
     * attendance group 班制 (groupType)
     *
     * - "1": 固定班制
     * - "2": 排班制
     * - "3": 自由工时
     */
    type AttendanceGroupType = '1' | '2' | '3';

    /** hrm attendance group (考勤组, HRM-12) */
    type AttendanceGroup = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** group name */
      groupName: string;
      /** group type: 固定班制/排班制/自由工时 */
      groupType: AttendanceGroupType | '';
      /** work days 1-7 (周一至周日) */
      workDays: number[];
      /** start time HH:mm (固定班制) */
      startTime: string;
      /** end time HH:mm (固定班制) */
      endTime: string;
      /** flexible minutes, 0 = not flexible */
      flexibleMinutes: number;
      /** 排班制未排班兜底班次 uuid (预留 HRM-13) */
      unassignedShiftUuid: string;
      /** 考勤组默认班次集合 (预留 HRM-13) */
      defaultShiftUuids: string[];
      /** 适用范围-部门 (含下级) */
      orgUuids: string[];
      /** 适用范围-指定人员 */
      memberUuids: string[];
      /** 负责人 */
      managers: string[];
      /** status: enabled/disabled */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** attendance group search params */
    type AttendanceGroupSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.AttendanceGroup, 'groupName' | 'groupType' | 'status'> & Common.CommonSearchParams
    >;

    /** attendance group paginated list */
    type AttendanceGroupList = Common.PaginatingQueryRecord<AttendanceGroup>;

    /** add attendance group request */
    type AddAttendanceGroupRequest = {
      groupName: string;
      groupType: AttendanceGroupType | '';
      workDays?: number[];
      startTime?: string;
      endTime?: string;
      flexibleMinutes?: number;
      unassignedShiftUuid?: string;
      defaultShiftUuids?: string[];
      orgUuids?: string[];
      memberUuids?: string[];
      managers?: string[];
      status?: AttendanceStatus;
      remark?: string;
    };

    /** edit attendance group request */
    type EditAttendanceGroupRequest = AddAttendanceGroupRequest & {
      uuid: string;
      status: AttendanceStatus;
    };

    /** effective group member (生效人数) */
    type EffectiveGroupMember = {
      uuid: string;
      employeeNo: string;
      name: string;
      orgName: string;
    };

    /** attendance group members response */
    type AttendanceGroupMembersResponse = {
      total: number;
      members: EffectiveGroupMember[];
    };

    /** effective attendance rule query response (三级回退: 考勤组 > 本部门 > 最近上级组织) */
    type EffectiveAttendanceRuleResponse = {
      found: boolean;
      /** rule source: "group"=考勤组 "org"=组织默认规则 */
      source: string;
      sourceUuid: string;
      sourceName: string;
      groupType: AttendanceGroupType | '';
      startTime: string;
      endTime: string;
      workDays: number[];
      flexibleMinutes: number;
    };

    /** shift usage response (班次引用查询, 支撑班次页 引用考勤组 列) */
    type AttendanceGroupShiftUsageResponse = {
      count: number;
      groups: {
        uuid: string;
        groupName: string;
        status: AttendanceStatus | null;
      }[];
    };

    /** hrm attendance config (组织默认考勤规则, 每组织一条; HRM-12) */
    type AttendanceConfig = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** org uuid */
      orgUuid: string;
      /** org name (derived) */
      orgName: string;
      startTime: string;
      endTime: string;
      /** work days 1-7 */
      workDays: number[];
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** attendance config search params */
    type AttendanceConfigSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.AttendanceConfig, 'orgName'> & Common.CommonSearchParams
    >;

    /** attendance config paginated list */
    type AttendanceConfigList = Common.PaginatingQueryRecord<AttendanceConfig>;

    /** save attendance config request (upsert: 快速开始/新增/编辑共用) */
    type SaveAttendanceConfigRequest = {
      orgUuid: string;
      startTime?: string;
      endTime?: string;
      workDays?: number[];
      remark?: string;
    };

    // ---- HRM 证书模板与证书发放 (HRM-30) ----

    /** certificate template status (证书模板状态) */
    type CertTemplateStatus = '1' | '2';

    /** certificate template source (模板来源) */
    type CertTemplateSource = '1' | '2';

    /** hrm training certificate template (证书模板) */
    type CertTemplate = {
      /** record uuid */
      uuid: string;
      /** 模板名称 */
      templateName: string;
      /** 证书类型（合格证/继续教育/荣誉/安全培训证/结业证/技能证明/上岗证/企业…） */
      certType: string;
      /** 来源 1=系统 2=企业 */
      source: CertTemplateSource;
      /** 编号前缀（HF/CE/YX/XF/RZ/JN/SG/AQ/HG/JX…），证书号=前缀-年-流水 */
      noPrefix: string;
      /** 状态 1=启用 2=停用（停用后不可发证） */
      status: CertTemplateStatus | null;
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** certificate template search params */
    type CertTemplateSearchParams = CommonType.RecordNullable<
      Pick<CertTemplate, 'templateName' | 'certType' | 'source' | 'status'> & Common.CommonSearchParams
    >;

    /** certificate template paginated list */
    type CertTemplateList = Common.PaginatingQueryRecord<CertTemplate>;

    /** add certificate template request */
    type AddCertTemplateRequest = {
      templateName: string;
      certType: string;
      source: CertTemplateSource;
      noPrefix: string;
      remark: string;
    };

    /** edit certificate template request */
    type EditCertTemplateRequest = AddCertTemplateRequest & {
      uuid: string;
    };

    /** toggle certificate template status request (1=启用 2=停用) */
    type ToggleCertTemplateStatusRequest = {
      uuid: string;
      status: CertTemplateStatus;
    };

    /** hrm training certificate (证书发放, admin view of hrm_training_certificate) */
    type TrainingCertificate = {
      /** record uuid */
      uuid: string;
      /** 证书号 前缀-年-流水（如 JN-2026-000001） */
      certNo: string;
      /** 证书名称 */
      certName: string;
      /** 证书类型 */
      certType: string;
      /** 证书模板 uuid */
      templateUuid: string;
      /** 模板名称（展示快照） */
      templateName: string;
      /** 来源 1=班次结业 2=外部导入 */
      source: TrainingCertificateSource;
      /** 来源班次 uuid */
      sessionUuid: string;
      /** 来源班次名称（展示快照） */
      sessionName: string;
      /** 学员（manage_user.uuid） */
      userUuid: string;
      /** 学员账号 */
      username: string;
      /** 学员姓名 */
      nickName: string;
      /** 发证时间 */
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

    /** training certificate (admin) search params */
    type TrainingCertificateSearchParams = CommonType.RecordNullable<
      Pick<TrainingCertificate, 'certNo' | 'certName' | 'certType' | 'status' | 'source' | 'sessionUuid'> &
        Common.CommonSearchParams
    >;

    /** training certificate (admin) paginated list */
    type TrainingCertificateList = Common.PaginatingQueryRecord<TrainingCertificate>;

    /** batch issue certificates request (班次结业批量发证) */
    type BatchIssueRequest = {
      sessionUuid: string;
      templateUuid: string;
      /** 缺省=班次全部学员 */
      userUuids?: string[];
    };

    /** batch issue certificates response */
    type BatchIssueResponse = {
      /** 本次实际新生成证书数（幂等跳过的不计） */
      issuedCount: number;
      /** 因已有有效证书被跳过的学员数 */
      skippedCount: number;
    };

    /** issue/revoke certificates request (批量 uuid) */
    type CertificateActionRequest = {
      uuids: string[];
    };

    /** import external certificate request (source='2', 证书号手动录入) */
    type ImportExternalCertificateRequest = {
      userUuid: string;
      certName: string;
      certType: string;
      certNo: string;
      fileName?: string;
      fileContent?: string;
    };

    /** upload certificate scan file request */
    type UploadCertificateFileRequest = {
      uuid: string;
      fileName: string;
      fileContent: string;
    };

    /** certificate scan file response: content encoded as base64 (mirrors 我的培训) */
    type DownloadCertificateFileResponse = {
      fileName: string;
      content: string;
    };

    type RewardType = 'reward' | 'punishment';

    /**
     * reward level (奖惩级别, demo-inventory §7)
     *
     * - "company": 公司级
     * - "dept": 部门级
     * - "notice": 通报
     * - "other": 其他
     */

    type RewardLevel = 'company' | 'dept' | 'notice' | 'other';

    /** hrm reward bill (奖惩单 HR215) */

    type HrmReward = {
      /** record uuid */
      uuid: string;
      /** 单据编号 HR215-... */
      billNo: string;
      /** 单据状态 */
      status: BillStatus | null;
      /** 员工档案 uuid */
      employeeUuid: string;
      /** 员工工号 */
      employeeNo: string;
      /** 员工姓名 */
      employeeName: string;
      /** 奖惩类型 */
      rewardType: RewardType | null;
      /** 奖惩级别 */
      rewardLevel: RewardLevel | null;
      /** 文号 */
      docNo: string;
      /** 发生日期 YYYY-MM-DD */
      occurDate: string;
      /** 金额（奖励为发放金额，处分/扣款为正数金额，0=不涉及） */
      amount: number;
      /** 奖惩原因/事由 */
      reason: string;
      /** 入账薪资年月 YYYY-MM（薪资数据源契约） */
      salaryMonth: string;
      /** 薪资数据源同步标记 1=已同步 */
      synced: number;
      /** 申请人 uuid */
      creatorUuid: string;
      /** 申请人姓名 */
      creatorName: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** reward search params */

    type RewardSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.HrmReward, 'billNo' | 'status' | 'rewardType' | 'employeeNo'> & Common.CommonSearchParams
    >;

    /** reward paginated list */

    type RewardList = Common.PaginatingQueryRecord<HrmReward>;

    /** add reward request */

    type AddRewardRequest = {
      employeeUuid: string;
      employeeNo: string;
      employeeName: string;
      rewardType: RewardType | null;
      rewardLevel: RewardLevel | null;
      docNo: string;
      occurDate: string;
      amount: number;
      reason: string;
      /** YYYY-MM，空=发生日期所在月 */
      salaryMonth: string;
    };

    /** edit reward request */

    type EditRewardRequest = AddRewardRequest & {
      uuid: string;
    };

    /**
     * requirement post category (岗位类别, demo-inventory §5)
     *
     * - "tech": 技术类
     * - "product": 产品类
     * - "design": 设计类
     * - "operation": 运营类
     * - "market": 市场类
     * - "function": 职能类
     * - "other": 其他
     */

    type RequirementPostCategory = 'tech' | 'product' | 'design' | 'operation' | 'market' | 'function' | 'other';

    /**
     * requirement urgency (紧急度, demo-inventory §5)
     *
     * - "normal": 普通
     * - "urgent": 紧急
     */

    type RequirementUrgency = 'normal' | 'urgent';

    /** recruitment requirement bill (招聘需求单 HR207) */

    type HrmRequirement = {
      /** record uuid */
      uuid: string;
      /** 单据编号 HR207-... */
      billNo: string;
      /** 单据状态 */
      status: BillStatus | null;
      /** 岗位名称 */
      postName: string;
      /** 岗位类别 */
      postCategory: RequirementPostCategory | null;
      /** 需求人数 */
      headcount: number;
      /** 紧急度 */
      urgency: RequirementUrgency | null;
      /** 期望到岗日期 YYYY-MM-DD */
      expectDate: string;
      /** 所属部门（组织 uuid） */
      orgUuid: string;
      /** 所属部门名称（快照） */
      orgName: string;
      /** 需求说明/任职要求 */
      requirementDesc: string;
      /** 申请人 uuid */
      creatorUuid: string;
      /** 申请人姓名 */
      creatorName: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** requirement search params */

    type RequirementSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.HrmRequirement, 'billNo' | 'status' | 'postName' | 'postCategory' | 'urgency' | 'orgUuid'> &
        Common.CommonSearchParams
    >;

    /** requirement paginated list */

    type RequirementList = Common.PaginatingQueryRecord<HrmRequirement>;

    /** add requirement request */

    type AddRequirementRequest = {
      postName: string;
      postCategory: RequirementPostCategory | null;
      headcount: number;
      urgency: RequirementUrgency | null;
      expectDate: string;
      orgUuid: string;
      orgName: string;
      requirementDesc: string;
    };

    /** edit requirement request */

    type EditRequirementRequest = AddRequirementRequest & {
      uuid: string;
    };

    /**
     * candidate source channel (来源渠道, demo-inventory §5)
     *
     * - "recommend": 内部推荐
     * - "website": 招聘网站
     * - "campus": 校园招聘
     * - "headhunter": 猎头
     * - "other": 其他
     */

    type CandidateSourceChannel = 'recommend' | 'website' | 'campus' | 'headhunter' | 'other';

    /**
     * candidate education (最高学历, demo-inventory §5)
     *
     * - "doctor": 博士
     * - "master": 硕士
     * - "bachelor": 本科
     * - "associate": 大专
     * - "senior": 高中
     * - "other": 其他
     */

    type CandidateEducation = 'doctor' | 'master' | 'bachelor' | 'associate' | 'senior' | 'other';

    /**
     * candidate status (候选人状态, demo-inventory §5)
     *
     * - "new": 新线索
     * - "delivered": 已投递
     * - "interviewing": 面试中
     * - "offer": 已发Offer
     * - "hired": 录用通过
     * - "rejected": 不通过
     */

    type CandidateStatus = 'new' | 'delivered' | 'interviewing' | 'offer' | 'hired' | 'rejected';

    /** recruitment candidate (候选人, HRM-26 候选人中心; 普通实体, 非单据底座) */

    type HrmCandidate = {
      /** record uuid */
      uuid: string;
      /** 姓名 */
      name: string;
      /** 手机号 */
      phone: string;
      /** 邮箱 */
      email: string;
      /** 来源渠道 */
      sourceChannel: CandidateSourceChannel | null;
      /** 最高学历 */
      education: CandidateEducation | null;
      /** 当前公司 */
      currentCompany: string;
      /** 当前岗位 */
      currentPosition: string;
      /** 期望薪资（自由文本，如 20k-30k） */
      expectSalary: string;
      /** 候选人状态 */
      status: CandidateStatus | null;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** candidate search params (demo-inventory §5: 筛选姓名/手机号/来源渠道) */

    type CandidateSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.HrmCandidate, 'name' | 'phone' | 'sourceChannel' | 'status'> & Common.CommonSearchParams
    >;

    /** candidate paginated list */

    type CandidateList = Common.PaginatingQueryRecord<HrmCandidate>;

    /** add candidate request (初始状态 新线索) */

    type AddCandidateRequest = {
      name: string;
      phone: string;
      email: string;
      sourceChannel: CandidateSourceChannel | null;
      education: CandidateEducation | null;
      currentCompany: string;
      currentPosition: string;
      expectSalary: string;
      remark: string;
    };

    /** edit candidate request (状态不在此接口修改) */

    type EditCandidateRequest = AddCandidateRequest & {
      uuid: string;
    };

    /** candidate status transition request (不可跳级/不可回退) */

    type UpdateCandidateStatusRequest = {
      uuid: string;
      status: CandidateStatus;
    };

    /** import candidates request (base64 xlsx, mirrors employee import) */
    type ImportCandidateRequest = {
      fileName: string;
      content: string;
    };

    type ImportCandidateResponse = {
      successCount: number;
    };

    /** 工作经历子表行 (入职单快照, demo-inventory §3.1) */
    type EntryBillWork = {
      uuid?: string;
      company: string;
      position: string;
      startDate: string;
      endDate: string;
      remark?: string;
    };

    /** 教育经历子表行 (入职单快照) */
    type EntryBillEducation = {
      uuid?: string;
      school: string;
      major: string;
      education: string;
      startDate: string;
      endDate: string;
      remark?: string;
    };

    /** 家庭成员子表行 (入职单快照) */
    type EntryBillFamily = {
      uuid?: string;
      name: string;
      relation: string;
      phone: string;
      workUnit: string;
      remark?: string;
    };

    /** 员工入职单 (HR201, demo-inventory §3.1) — 审批通过自动建员工档案 */
    type HrmEntryBill = {
      /** record uuid */
      uuid: string;
      /** 单据编号 HR201-... */
      billNo: string;
      /** 单据状态 */
      status: BillStatus | null;
      /** 姓名 */
      name: string;
      /** 手机号 */
      phone: string;
      /** 邮箱 */
      email: string;
      /** 性别 1=男 2=女 */
      gender: string;
      /** 出生日期 YYYY-MM-DD */
      birthday: string;
      /** 民族 */
      nation: string;
      /** 政治面貌 */
      politicalStatus: string;
      /** 婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶 */
      maritalStatus: string;
      /** 文化程度 */
      cultureLevel: string;
      /** 籍贯 */
      nativePlace: string;
      /** 现居住地 */
      address: string;
      /** 户口所在地 */
      householdAddress: string;
      /** 紧急联系人 */
      emergencyContact: string;
      /** 紧急联系人电话 */
      emergencyPhone: string;
      /** 工资开户行 */
      bankName: string;
      /** 工资卡账户 */
      bankAccount: string;
      /** 员工照片 */
      avatar: string;
      /** 职位 */
      position: string;
      /** 职务 */
      post: string;
      /** 试用期(月) */
      probationMonths: number;
      /** 所属部门 (组织 uuid) */
      orgUuid: string;
      /** 所属公司 (组织 uuid) */
      companyUuid: string;
      /** 入职日期 YYYY-MM-DD */
      hireDate: string;
      /** 预计转正日期 YYYY-MM-DD */
      expectFormalDate: string;
      /** 附件说明 */
      attachmentDesc: string;
      /** 备注 */
      remark: string;
      /** 审批通过后创建的员工档案 uuid (空=未建档) */
      employeeUuid: string;
      /** 申请人 uuid */
      creatorUuid: string;
      /** 申请人姓名 */
      creatorName: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
      /** 工作经历快照 (详情返回) */
      works?: EntryBillWork[];
      /** 教育经历快照 (详情返回) */
      educations?: EntryBillEducation[];
      /** 家庭成员快照 (详情返回) */
      families?: EntryBillFamily[];
    };

    /** entry search params (demo-inventory §3.1: 单据编号/单据状态/姓名) */
    type EntrySearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.HrmEntryBill, 'billNo' | 'status' | 'name'> & Common.CommonSearchParams
    >;

    /** entry paginated list */
    type EntryList = Common.PaginatingQueryRecord<HrmEntryBill>;

    /** add entry request (预计转正空=入职日期+试用期月, 后端计算) */
    type AddEntryRequest = {
      name: string;
      phone: string;
      email: string;
      gender: string;
      birthday: string;
      nation: string;
      politicalStatus: string;
      maritalStatus: string;
      cultureLevel: string;
      nativePlace: string;
      address: string;
      householdAddress: string;
      emergencyContact: string;
      emergencyPhone: string;
      bankName: string;
      bankAccount: string;
      avatar: string;
      position: string;
      post: string;
      probationMonths: number;
      orgUuid: string;
      companyUuid: string;
      /** YYYY-MM-DD */
      hireDate: string;
      /** YYYY-MM-DD, 空=入职日期+试用期月 */
      expectFormalDate: string;
      attachmentDesc: string;
      remark: string;
      works: EntryBillWork[];
      educations: EntryBillEducation[];
      families: EntryBillFamily[];
    };

    /** edit entry request (only in editable 单据状态) */
    type EditEntryRequest = AddEntryRequest & {
      uuid: string;
    };

    /** entry detail (带三子表快照) */
    type EntryDetail = {
      entry: HrmEntryBill;
    };

    // ===== HRM-33 薪酬主体与薪资项目 =====

    /** 薪酬主体 status: 1 enabled 0 disabled */
    type SalarySubjectStatus = '1' | '0';

    /** hrm salary subject (薪酬主体, GLOSSARY: 发薪的法人实体, 关联一个公司) */
    type SalarySubject = {
      /** record uuid */
      uuid: string;
      /** subject name */
      subjectName: string;
      /** linked company uuid (hrm_org org_type=1) */
      orgUuid: string;
      /** linked company name snapshot */
      orgName: string;
      /** pay account */
      payAccount: string;
      /** pay bank */
      payBank: string;
      /** status: enabled/disabled */
      status: SalarySubjectStatus | null;
      /** remark */
      remark: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** salary subject search params */
    type SalarySubjectSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.SalarySubject, 'subjectName' | 'status'> & Common.CommonSearchParams
    >;

    /** salary subject paginated list */
    type SalarySubjectList = Common.PaginatingQueryRecord<SalarySubject>;

    /** add salary subject request */
    type AddSalarySubjectRequest = {
      subjectName: string;
      orgUuid: string;
      payAccount: string;
      payBank: string;
      status: SalarySubjectStatus | null;
      remark: string;
    };

    /** edit salary subject request */
    type EditSalarySubjectRequest = {
      uuid: string;
      subjectName: string;
      orgUuid: string;
      payAccount: string;
      payBank: string;
      status: SalarySubjectStatus | null;
      remark: string;
    };

    /** 薪资项目类型: fixed=固定项 data=取数项 formula=公式项 */
    type SalaryItemType = 'fixed' | 'data' | 'formula';

    /** 薪资项目分类: earning=应发 deduction=应扣 cost=成本 */
    type SalaryItemClass = 'earning' | 'deduction' | 'cost';

    /** 取值方式: manual=手工维护 system=系统取数 formula=公式计算 */
    type SalaryItemValueMode = 'manual' | 'system' | 'formula';

    /** hrm salary item (薪资项目, GLOSSARY: 薪资计算的最小条目) */
    type SalaryItem = {
      /** record uuid */
      uuid: string;
      /** item code (BASIC_SALARY/NET_SALARY…) */
      itemCode: string;
      /** item name */
      itemName: string;
      /** item type */
      itemType: SalaryItemType;
      /** item class */
      itemClass: SalaryItemClass;
      /** value mode */
      valueMode: SalaryItemValueMode;
      /** calc order (试算按此升序消费, 同序按 id) */
      calcOrder: number;
      /** display order */
      orderNum: number;
      /** formula expression (formula items only) */
      formula: string;
      /** preset flag: 1 preset (不可删除) */
      isPreset: number;
      /** status: enabled/disabled */
      status: SalarySubjectStatus | null;
      /** remark */
      remark: string;
      /** record create time */
      createTime: string;
      /** record update time */
      updateTime: string;
    };

    /** salary item search params */
    type SalaryItemSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.SalaryItem, 'itemName' | 'itemClass' | 'status'> & Common.CommonSearchParams
    >;

    /** salary item paginated list */
    type SalaryItemList = Common.PaginatingQueryRecord<SalaryItem>;

    /** add salary item request */
    type AddSalaryItemRequest = {
      itemCode: string;
      itemName: string;
      itemType: SalaryItemType;
      itemClass: SalaryItemClass;
      valueMode: SalaryItemValueMode;
      calcOrder: number;
      orderNum: number;
      formula: string;
      status: SalarySubjectStatus | null;
      remark: string;
    };

    /** edit salary item request */
    type EditSalaryItemRequest = {
      uuid: string;
      itemCode: string;
      itemName: string;
      itemType: SalaryItemType;
      itemClass: SalaryItemClass;
      valueMode: SalaryItemValueMode;
      calcOrder: number;
      orderNum: number;
      formula: string;
      status: SalarySubjectStatus | null;
      remark: string;
    };

    /**
     * overtime day type (当天类型)
     *
     * - "1": 工作日
     * - "2": 休息日
     * - "3": 法定节假日
     */
    type OvertimeDayType = '1' | '2' | '3';

    /** overtime policy of one day type (当天策略: 审批/打卡/调休倍数) */
    type DayOvertimePolicy = {
      /** 需审批 1=需审批 0=免审批 */
      approvalRequired: number;
      /** 需打卡 1=需打卡 0=免打卡 */
      punchRequired: number;
      /** 调休倍数 (1=1:1) */
      compMultiplier: number;
    };

    /** hrm overtime rule (加班规则, HRM-13) */
    type OvertimeRule = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 规则名称 */
      ruleName: string;
      /** 工作日策略 */
      workday: DayOvertimePolicy;
      /** 休息日策略 */
      restday: DayOvertimePolicy;
      /** 法定节假日策略 */
      holiday: DayOvertimePolicy;
      /** 最小计算单位（分钟, 默认30） */
      minUnitMinutes: number;
      /** 月上限（小时, 0=不限制） */
      monthlyLimitHours: number;
      /** 使用的考勤组 uuid 集合 (空=全局兜底规则) */
      groupUuids: string[];
      /** 组织默认规则所属组织 uuid (空=全局) */
      orgUuid: string;
      /** status: enabled/disabled */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** overtime rule search params */
    type OvertimeRuleSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.OvertimeRule, 'ruleName' | 'status'> & Common.CommonSearchParams
    >;

    /** overtime rule paginated list */
    type OvertimeRuleList = Common.PaginatingQueryRecord<OvertimeRule>;

    /** add overtime rule request */
    type AddOvertimeRuleRequest = {
      ruleName: string;
      workday: DayOvertimePolicy;
      restday: DayOvertimePolicy;
      holiday: DayOvertimePolicy;
      minUnitMinutes?: number;
      monthlyLimitHours?: number;
      groupUuids?: string[];
      orgUuid?: string;
      status?: AttendanceStatus;
      remark?: string;
    };

    /** edit overtime rule request */
    type EditOvertimeRuleRequest = AddOvertimeRuleRequest & {
      uuid: string;
      status: AttendanceStatus;
    };

    /** hrm schedule cycle (排班周期, HRM-14): 循环使用的班次序列, 周期天数 = 序列长度（派生） */
    type ScheduleCycle = {
      /** record uuid */
      uuid: string;
      /** record creator */
      createBy: string;
      /** record updater */
      updateBy: string;
      /** 周期名称 */
      cycleName: string;
      /** 有序班次 uuid 序列（循环展开） */
      shiftUuids: string[];
      /** 适用考勤组 uuid 集合（空=未绑定） */
      groupUuids: string[];
      /** status: enabled/disabled */
      status: AttendanceStatus | null;
      /** remark */
      remark: string;
      createTime: string;
      updateTime: string;
    };

    /** schedule cycle search params */
    type ScheduleCycleSearchParams = CommonType.RecordNullable<
      Pick<Api.Hrm.ScheduleCycle, 'cycleName' | 'status'> & Common.CommonSearchParams
    >;

    /** schedule cycle paginated list */
    type ScheduleCycleList = Common.PaginatingQueryRecord<ScheduleCycle>;

    /** add schedule cycle request */
    type AddScheduleCycleRequest = {
      cycleName: string;
      shiftUuids: string[];
      groupUuids?: string[];
      status?: AttendanceStatus;
      remark?: string;
    };

    /** edit schedule cycle request */
    type EditScheduleCycleRequest = AddScheduleCycleRequest & {
      uuid: string;
      status: AttendanceStatus;
    };
  }
}
