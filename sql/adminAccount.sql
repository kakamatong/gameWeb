CREATE TABLE `adminAccount` (
  -- 核心ID
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '账户唯一ID，主键',
  
  -- 账户身份信息
  `username` varchar(50) NOT NULL COMMENT '登录用户名，唯一',
  `passwordHash` varchar(255) NOT NULL COMMENT '加密后的密码',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱，可用于登录或找回密码',
  `mobile` varchar(20) DEFAULT NULL COMMENT '手机号，可用于登录或通知',
  
  -- 账户状态控制
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '账户状态：0-禁用，1-启用',
  `isSuperAdmin` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是超级管理员：0-否，1-是',
  
  -- 个人信息
  `realName` varchar(50) DEFAULT NULL COMMENT '用户真实姓名',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL地址',
  `departmentId` int(11) DEFAULT NULL COMMENT '所属部门ID',
  `note` varchar(500) DEFAULT NULL COMMENT '备注信息',
  
  -- 安全与登录信息
  `lastLoginIp` varchar(45) DEFAULT NULL COMMENT '最后一次登录IP',
  `lastLoginTime` datetime DEFAULT NULL COMMENT '最后一次登录时间',
  `passwordResetToken` varchar(255) DEFAULT NULL COMMENT '密码重置令牌',
  `tokenExpireTime` datetime DEFAULT NULL COMMENT '令牌过期时间',
  
  -- 审计日志字段 (创建/修改信息)
  `createdBy` bigint(20) UNSIGNED DEFAULT NULL COMMENT '创建者ID',
  `updatedBy` bigint(20) UNSIGNED DEFAULT NULL COMMENT '最后修改者ID',
  `createdTime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updatedTime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  
  -- 设置主键和唯一约束
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_mobile` (`mobile`),
  KEY `idx_department_id` (`departmentId`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理系统账户表';