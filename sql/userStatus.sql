CREATE TABLE `userStatus` (
  `userid` bigint NOT NULL COMMENT '用户唯一标识',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '玩家状态（0:离线,1:大厅,2:匹配中,3:准备中,4:游戏中,5:观战,6:组队中,7:断线）',
  `gameid` bigint DEFAULT 0 COMMENT '当前所在游戏房间/对局的ID（非游戏状态时为NULL）',
  `roomid` bigint DEFAULT 0 COMMENT '当前所在房间ID（非游戏状态时为NULL）',
  `shortRoomid` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '私有房id',
  `addr` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '地址',
  `ext` text COLLATE utf8mb4_unicode_ci COMMENT '扩展字段(JSON格式存储)',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`userid`),
  KEY `idx_roomid` (`roomid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='玩家实时状态表'