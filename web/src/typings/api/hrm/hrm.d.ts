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
  }
}
