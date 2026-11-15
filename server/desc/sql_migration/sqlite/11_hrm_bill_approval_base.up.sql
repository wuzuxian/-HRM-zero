-- HRM-06 单据审批底座（ADR-0001）：共享审批记录表 + 单据编号流水表
DROP TABLE IF EXISTS hrm_approval_record;

CREATE TABLE hrm_approval_record (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_type_code TEXT NOT NULL,
    bill_uuid TEXT NOT NULL,
    bill_no TEXT NOT NULL,
    "action" TEXT NOT NULL,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    operator_user_uuid TEXT NOT NULL,
    operator_name TEXT NOT NULL DEFAULT '',
    opinion TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_approval_record_bill ON hrm_approval_record (bill_type_code, bill_uuid);

DROP TABLE IF EXISTS hrm_bill_sequence;

CREATE TABLE hrm_bill_sequence (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_type_code TEXT NOT NULL,
    bill_date TEXT NOT NULL,
    last_seq INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX uni_hrm_bill_sequence_type_date ON hrm_bill_sequence (bill_type_code, bill_date);

CREATE TRIGGER hrm_approval_record_update_time
AFTER UPDATE ON hrm_approval_record
FOR EACH ROW
BEGIN
    UPDATE hrm_approval_record SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_bill_sequence_update_time
AFTER UPDATE ON hrm_bill_sequence
FOR EACH ROW
BEGIN
    UPDATE hrm_bill_sequence SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
