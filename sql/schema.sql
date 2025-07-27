# 文件路径: sql/schema.sql

-- 创建 links 表用于存储短链接和原始链接的映射关系
CREATE TABLE IF NOT EXISTS `links` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `created_at` DATETIME(3) NULL COMMENT '创建时间',
    `original_url` VARCHAR(2048) NOT NULL COMMENT '原始URL',
    `short_code` VARCHAR(10) NULL COMMENT '生成的短代码',
    UNIQUE INDEX `idx_links_short_code` (`short_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;