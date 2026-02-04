-- WellCMS Go User Table
-- 用户表

CREATE TABLE IF NOT EXISTS `user` (
  `uid` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `username` VARCHAR(32) NOT NULL COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '加密后的密码',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `role` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色: 0-普通用户, 1-管理员',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0-正常, 1-禁用',
  `dateline` INT UNSIGNED NOT NULL COMMENT '注册时间',
  `lastvisit` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后访问时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 插入默认管理员用户
-- 密码: admin123 (bcrypt加密)
INSERT INTO `user` (`uid`, `username`, `password`, `email`, `avatar`, `role`, `status`, `dateline`, `lastvisit`) VALUES
(1, 'admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'admin@example.com', '', 1, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();
