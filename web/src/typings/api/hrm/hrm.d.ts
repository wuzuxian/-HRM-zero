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
  }
}
