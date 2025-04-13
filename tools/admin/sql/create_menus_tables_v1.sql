-- 生成时间: 2025-04-13 22:23:07
-- 模块: admin

-- 菜单表
CREATE TABLE IF NOT EXISTS menus (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  parent_id INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父级ID',
  path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '地址',
  title VARCHAR(100) NOT NULL DEFAULT '' COMMENT '标题',
  name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '路由中的name',
  component VARCHAR(255) NOT NULL DEFAULT '' COMMENT '绑定的组件，默认类型：Iframe、RouteView、ComponentError',
  locale VARCHAR(100) NOT NULL DEFAULT '' COMMENT '本地化标识',
  icon VARCHAR(100) NOT NULL DEFAULT '' COMMENT '图标',
  redirect VARCHAR(255) NOT NULL DEFAULT '' COMMENT '重定向地址',
  url VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'iframe模式下的跳转url，不能与path重复',
  keep_alive TINYINT(3) NOT NULL DEFAULT 1 COMMENT '是否缓存',
  hide_menu TINYINT(3) NOT NULL DEFAULT 1 COMMENT '是否隐藏',
  target VARCHAR(20) NOT NULL DEFAULT '' COMMENT '全连接跳转模式',
  weight INT NOT NULL DEFAULT 0 COMMENT '排序权重',
  created_at DATETIME COMMENT '创建时间',
  updated_at DATETIME COMMENT '更新时间',
  deleted_at DATETIME COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY `menus_parent_id_idx` (parent_id) COMMENT '父级ID索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单表';

