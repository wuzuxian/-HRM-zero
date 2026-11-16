// 共享单据导出（final-review F5, spec 用户故事 23: 每类人事单据支持导出）。
// 四个人事管理列表页（entry/regular/transfer/resignation）共用：调
// /api/v1/hrm/exportBillList（与列表同一套 mine 过滤口径），下载返回的 base64 xlsx。
// 按钮码 v1:hrm:bill:exportBillList（种子：迁移 56，人事管理目录下）。
import { ExportBillList } from '@/service/api';

export function useBillExport(billType: string) {
  async function exportBills() {
    const { error, data: exported } = await ExportBillList({ billType });
    if (error || !exported) return;
    const bin = atob(exported.content);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i += 1) {
      bytes[i] = bin.charCodeAt(i);
    }
    const blob = new Blob([bytes], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = exported.fileName;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  return { exportBills };
}
