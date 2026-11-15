-- HRM-06 单据审批底座（ADR-0001）：共享审批记录表 + 单据编号流水表
DROP TABLE IF EXISTS "hrm_approval_record";

CREATE TABLE "hrm_approval_record" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_type_code varchar(8) NOT NULL,
    bill_uuid varchar(36) NOT NULL,
    bill_no varchar(32) NOT NULL,
    "action" varchar(16) NOT NULL,
    from_status varchar(16) NOT NULL,
    to_status varchar(16) NOT NULL,
    operator_user_uuid varchar(36) NOT NULL,
    operator_name varchar(64) NOT NULL DEFAULT '',
    opinion varchar(255) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_approval_record_bill ON "hrm_approval_record" (bill_type_code, bill_uuid);

CREATE TRIGGER update_hrm_approval_record_update_time BEFORE UPDATE ON "hrm_approval_record"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

DROP TABLE IF EXISTS "hrm_bill_sequence";

CREATE TABLE "hrm_bill_sequence" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_type_code varchar(8) NOT NULL,
    bill_date char(8) NOT NULL,
    last_seq bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uni_hrm_bill_sequence_type_date ON "hrm_bill_sequence" (bill_type_code, bill_date);

CREATE TRIGGER update_hrm_bill_sequence_update_time BEFORE UPDATE ON "hrm_bill_sequence"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();
