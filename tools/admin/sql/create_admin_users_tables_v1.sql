-- 生成时间: 2025-04-13 22:22:49
-- 模块: admin

-- 管理员用户表
CREATE TABLE IF NOT EXISTS admin_users (
  ID INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用户名',
  email VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  password VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
  status TINYINT(3) NOT NULL DEFAULT 1 COMMENT '状态',
  created_at DATETIME COMMENT '创建时间',
  updated_at DATETIME COMMENT '更新时间',
  deleted_at DATETIME COMMENT '删除时间',
  PRIMARY KEY (ID),
  UNIQUE KEY `admin_users_name_unique` (name) COMMENT '用户名唯一索引',
  UNIQUE KEY `admin_users_email_unique` (email) COMMENT '邮箱唯一索引',
  KEY `admin_users_status_idx` (status) COMMENT '状态索引',
  KEY `admin_users_status_created_idx` (status,created_at) COMMENT '状态和创建时间复合索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员用户表';

