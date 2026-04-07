-- 创建数据库（如果 docker-compose 没创建的话）
CREATE DATABASE IF NOT EXISTS `gonetdisk` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `gonetdisk`;

-- 1. 用户表
CREATE TABLE `user` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `email` VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `avatar_url` VARCHAR(255) COMMENT '头像地址',
    `used_space` BIGINT UNSIGNED DEFAULT 0 COMMENT '已用空间(Byte)',
    `total_space` BIGINT UNSIGNED DEFAULT 1073741824 COMMENT '总空间(默认1GB)',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0正常, 1禁用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_username` (`username`)
) ENGINE = InnoDB;

-- 2. 物理文件表 (核心：去重与秒传的基础)
CREATE TABLE `physical_file` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `file_hash` CHAR(64) NOT NULL UNIQUE COMMENT '文件唯一哈希(如md5)',
    `file_name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
    `file_ext` VARCHAR(20) COMMENT '文件扩展名',
    `file_size` BIGINT UNSIGNED NOT NULL COMMENT '文件实际大小',
    `file_path` VARCHAR(500) NOT NULL COMMENT '物理存储路径',
    `storage_type` VARCHAR(20) DEFAULT 'local' COMMENT '存储方法: local, oss, s3',
    `ref_count` INT UNSIGNED DEFAULT 1 COMMENT '引用计数',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_hash` (`file_hash`)
) ENGINE = InnoDB;

-- 3. 用户文件表 (逻辑层级)
CREATE TABLE `user_file` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '所属用户ID',
    `physical_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联物理文件ID(目录则为空)',
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父文件夹ID, 0为根目录',
    `file_name` VARCHAR(255) NOT NULL COMMENT '用户显示的文件名',
    `file_ext` VARCHAR(20) COMMENT '扩展名',
    `file_size` BIGINT UNSIGNED NOT NULL COMMENT '文件实际大小',
    `path_stack` TEXT COMMENT '族谱路径, 如 /0/1/5',
    `is_dir` TINYINT(1) DEFAULT 0 COMMENT '是否为文件夹: 1是, 0否',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_user_parent` (`user_id`, `parent_id`),
    INDEX `idx_physical` (`physical_id`)
) ENGINE = InnoDB;

-- 4. 角色表
CREATE TABLE `role` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `role_name` VARCHAR(50) NOT NULL COMMENT '角色名',
    `role_key` VARCHAR(50) NOT NULL UNIQUE COMMENT '角色唯一标识',
    `description` VARCHAR(255) COMMENT '角色描述',
    `status` TINYINT DEFAULT 1 COMMENT '状态',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL
) ENGINE = InnoDB;

-- 5. 管理员表
CREATE TABLE `admin` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '管理员名',
    `password` VARCHAR(255) NOT NULL COMMENT '密码',
    `role_id` INT UNSIGNED NOT NULL COMMENT '角色ID',
    `last_login` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_role` (`role_id`)
) ENGINE = InnoDB;

-- 6. 权限表
CREATE TABLE `permission` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `parent_id` INT UNSIGNED DEFAULT 0 COMMENT '父级ID',
    `perm_name` VARCHAR(50) NOT NULL COMMENT '权限名称',
    `perm_key` VARCHAR(100) NOT NULL COMMENT '权限标识',
    `type` TINYINT COMMENT '类型: 1菜单, 2按钮, 3接口',
    `api_path` VARCHAR(255) COMMENT 'API路径',
    `http_method` VARCHAR(10) COMMENT 'HTTP方法',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB;

-- 7. 角色-权限关联表
CREATE TABLE `role_permission` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `role_id` INT UNSIGNED NOT NULL,
    `permission_id` INT UNSIGNED NOT NULL,
    UNIQUE KEY `uk_role_perm` (`role_id`, `permission_id`)
) ENGINE = InnoDB;