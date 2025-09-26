CREATE TABLE `account` (
  `username` char(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户唯一标识',
  `userid` bigint NOT NULL AUTO_INCREMENT,
  `password` char(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码（MD5等32位哈希存储）',
  `type` char(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'account' COMMENT '账户类型',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`username`),
  UNIQUE KEY `uniq_numid` (`userid`),
  KEY `idx_userid_password` (`username`,`password`)
) ENGINE=InnoDB AUTO_INCREMENT=10003 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 