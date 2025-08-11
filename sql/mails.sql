CREATE TABLE mails (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '邮件唯一ID',
    type INT NOT NULL COMMENT '邮件类型: 0-全服邮件, 1-个人邮件',
    senderid BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID, 0表示系统',
    title VARCHAR(100) NOT NULL COMMENT '邮件标题',
    content TEXT COMMENT '邮件内容',
    awards VARCHAR(512) COMMENT '奖励内容，通常为JSON格式',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    -- 索引
    INDEX idx_senderid (senderid) COMMENT '发送者ID索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮件表';