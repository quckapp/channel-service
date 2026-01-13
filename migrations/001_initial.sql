CREATE TABLE IF NOT EXISTS channels (
    id CHAR(36) PRIMARY KEY,
    workspace_id CHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('public', 'private', 'dm', 'group_dm') DEFAULT 'public',
    description TEXT,
    topic VARCHAR(500),
    icon_url VARCHAR(500),
    is_archived BOOLEAN DEFAULT FALSE,
    created_by CHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_channels_workspace (workspace_id),
    INDEX idx_channels_type (type),
    UNIQUE KEY unique_channel_name (workspace_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS channel_members (
    id CHAR(36) PRIMARY KEY,
    channel_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    role ENUM('owner', 'admin', 'member') DEFAULT 'member',
    notifications ENUM('all', 'mentions', 'none') DEFAULT 'all',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_read_at TIMESTAMP NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    UNIQUE KEY unique_channel_member (channel_id, user_id),
    INDEX idx_channel_members_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS channel_pins (
    id CHAR(36) PRIMARY KEY,
    channel_id CHAR(36) NOT NULL,
    message_id CHAR(36) NOT NULL,
    pinned_by CHAR(36) NOT NULL,
    pinned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    UNIQUE KEY unique_pin (channel_id, message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
