CREATE TABLE mailUsers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '序号，主键',
    userid BIGINT NOT NULL COMMENT '用户ID',
    mailid BIGINT NOT NULL COMMENT '邮件ID，关联mails表id',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-未读, 1-已读, 2-已领取, 3-已删除',
    startTime DATETIME NOT NULL COMMENT '邮件生效开始时间',
    endTime DATETIME NOT NULL COMMENT '邮件生效结束时间',
    update_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
    
    -- 基础索引
    INDEX idx_userid (userid) COMMENT '用户ID索引',
    INDEX idx_user_mail (userid, mailid) COMMENT '用户和邮件组合索引',
    INDEX idx_startTime (startTime) COMMENT '开始时间索引',
    INDEX idx_endTime (endTime) COMMENT '结束时间索引',
    
    -- 高级优化索引
    INDEX idx_user_status (userid, status) COMMENT '用户和状态组合索引',
    INDEX idx_user_time (userid, endTime) COMMENT '用户和结束时间组合索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='玩家邮件表';