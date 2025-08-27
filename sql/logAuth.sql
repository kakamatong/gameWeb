CREATE TABLE `logAuth` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `userid` bigint NOT NULL,
  `nickname` varchar(64) NOT NULL,
  `ip` varchar(50) DEFAULT NULL,
  `loginType` varchar(32) DEFAULT NULL COMMENT '认证类型（渠道）',
  `status` tinyint(1) DEFAULT NULL COMMENT '认证状态(0失败 1成功)',
  `ext` varchar(256) DEFAULT NULL COMMENT '扩展数据',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`userid`),
  KEY `idx_status` (`status`),
  KEY `idx_loginType` (`loginType`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证日志表';