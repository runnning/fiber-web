-- 生成时间: 2025-04-13 22:22:58
-- 模块: admin

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
  ID INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '角色名',
  created_at DATETIME COMMENT '创建时间',
  updated_at DATETIME COMMENT '更新时间',
  deleted_at DATETIME COMMENT '删除时间',
  PRIMARY KEY (ID),
  UNIQUE KEY `roles_name_unique` (name) COMMENT '角色名唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

