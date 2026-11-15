import { request } from '../request';

/** get hrm org tree */
export function GetOrgTree(params?: Api.Hrm.OrgSearchParams) {
  return request<Api.Hrm.OrgTree>({
    url: '/api/v1/hrm/getOrgTree',
    method: 'get',
    params
  });
}

/** add hrm org (root or with parent) */
export function AddOrg(req: Api.Hrm.AddOrgRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addOrg',
    method: 'post',
    data: req
  });
}

/** add a child hrm org under an existing one */
export function AddOrgChild(req: Api.Hrm.AddOrgChildRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addOrgChild',
    method: 'post',
    data: req
  });
}

/** edit hrm org */
export function EditOrg(req: Api.Hrm.EditOrgRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editOrg',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm orgs */
export function DeleteOrg(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteOrg',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** import hrm orgs from a base64-encoded xlsx file */
export function ImportOrg(req: Api.Hrm.ImportOrgRequest) {
  return request<Api.Hrm.ImportOrgResponse>({
    url: '/api/v1/hrm/importOrg',
    method: 'post',
    data: req
  });
}

/** export hrm orgs as a base64-encoded xlsx file */
export function ExportOrg() {
  return request<Api.Hrm.ExportOrgResponse>({
    url: '/api/v1/hrm/exportOrg',
    method: 'get'
  });
}
