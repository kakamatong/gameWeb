CREATE TABLE `userRiches` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键ID',
  `userid` bigint NOT NULL COMMENT '用户ID',
  `richType` int NOT NULL DEFAULT '0' COMMENT '财富类型',
  `richNums` bigint NOT NULL DEFAULT '0' COMMENT '财富数量',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_userid_richType` (`userid`, `richType`) COMMENT '用户ID和财富类型唯一组合',
  KEY `idx_userid` (`userid`) COMMENT '用户ID索引'
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户财富表';