CREATE TABLE mailSystem (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '序号，主键',
    type INT NOT NULL COMMENT '邮件类型: 0-全服邮件, 1-个人邮件, 2-按渠道邮件, 3-按登入类型邮件',
    mailid BIGINT NOT NULL COMMENT '邮件ID，关联mails表id',
    startTime DATETIME NOT NULL COMMENT '邮件生效开始时间',
    endTime DATETIME NOT NULL COMMENT '邮件生效结束时间',
    
    -- 索引
    INDEX idx_type (type) COMMENT '邮件类型索引',
    INDEX idx_mailid (mailid) COMMENT '邮件ID索引',
    INDEX idx_startTime (startTime) COMMENT '开始时间索引',
    INDEX idx_endTime (endTime) COMMENT '结束时间索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统邮件表';