# VideoWeb - 高性能视频后端服务 (Practice Project)

一个基于 Go 语言、Hertz 框架和 Thrift IDL 开发的轻量级视频网站后端程序。本项目作为练手项目，实现了用户系统、视频管理、社交关系及互动评论等核心功能。

## 🚀 技术栈

- **Web 框架**: [CloudWeGo/Hertz](https://github.com/cloudwego/hertz) (高性能 Go HTTP 框架)
- **IDL & RPC**: [Thrift](https://thrift.apache.org/) (接口定义与序列化)
- **数据库 ORM**: [GORM](https://gorm.io/) (v1)
- **存储**:
  - **MySQL**: 核心业务数据持久化
  - **Redis**: 缓存与状态管理
- **认证**: [JWT](https://github.com/hertz-contrib/jwt) (基于 hertz-contrib/jwt)
- **部署**: Docker 容器化

## ✨ 核心功能

- **用户系统**: 注册、登录、个人主页信息管理。
- **视频模块**: 视频上传、搜索、热门排行、用户视频列表（视频文件存储于本地服务器）。
- **互动系统**: 视频点赞、评论（支持二级评论/父子嵌套结构）。
- **社交系统**: 关注与粉丝关系管理。

## 📁 项目目录结构

```text
.
├── biz/                # 业务逻辑层
│   ├── dal/            # 数据访问层 (MySQL, Redis)
│   ├── handler/        # 请求处理器 (接口具体实现)
│   ├── model/          # Thrift 生成的数据模型
│   ├── my_jwt/         # JWT 认证中间件配置
│   └── router/         # 路由注册与中间件配置
├── idl/                # Thrift 接口定义文件 (.thrift)
├── main.go             # 服务入口
├── router.go           # 路由入口
├── go.mod              # 依赖管理
└── README.md           # 项目文档
```

## 🗄️ 数据库设计 (Database Schema)

项目数据库名为 `video_web`，核心表结构及其关系如下：

### 1. 数据库初始化 SQL
你可以直接复制并运行以下 SQL 脚本来完成数据库初始化：

```sql
CREATE DATABASE video_web; 
USE video_web;

-- 用户表
CREATE TABLE users ( 
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID', 
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名', 
    PASSWORD VARCHAR(255) NOT NULL COMMENT 'hash加密后密码', 
    avatar_url VARCHAR(500) COMMENT '头像URL', 
    nickname VARCHAR(50) NOT NULL COMMENT '昵称', 
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', 
    deleted_at DATETIME COMMENT '软删除时间（NULL表示未删除）', 
    INDEX idx_deleted_at (deleted_at) 
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='用户表'; 

-- 视频表
CREATE TABLE videos ( 
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '视频ID', 
    user_id BIGINT NOT NULL COMMENT '作者ID', 
    title VARCHAR(100) NOT NULL COMMENT '视频标题', 
    DESCRIPTION VARCHAR(500) COMMENT '视频描述', 
    video_url VARCHAR(500) NOT NULL COMMENT '视频URL', 
    views BIGINT DEFAULT 0 COMMENT '播放量', 
    likes BIGINT DEFAULT 0 COMMENT '点赞数', 
    comments BIGINT DEFAULT 0 COMMENT '评论数', 
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', 
    deleted_at DATETIME COMMENT '软删除时间（NULL表示未删除）', 
    INDEX idx_user_id (user_id), 
    INDEX idx_created_at (created_at), 
    INDEX idx_views (views), 
    INDEX idx_deleted_at (deleted_at), 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE 
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='视频表'; 

-- 评论表
CREATE TABLE comments ( 
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '评论ID', 
    user_id BIGINT NOT NULL COMMENT '评论用户ID', 
    video_id BIGINT NOT NULL COMMENT '视频ID', 
    parent_id BIGINT DEFAULT 0 COMMENT '父评论ID（0表示一级评论，否则为回复）', 
    content VARCHAR(500) NOT NULL COMMENT '评论内容', 
    likes BIGINT DEFAULT 0 COMMENT '点赞数', 
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', 
    deleted_at DATETIME COMMENT '软删除时间（NULL表示未删除）', 
    INDEX idx_user_id (user_id), 
    INDEX idx_video_id (video_id), 
    INDEX idx_parent_id (parent_id), 
    INDEX idx_created_at (created_at), 
    INDEX idx_deleted_at (deleted_at), 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE, 
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE 
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='评论表'; 

-- 点赞表
CREATE TABLE likes ( 
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '点赞ID', 
    user_id BIGINT NOT NULL COMMENT '用户ID', 
    target_id BIGINT NOT NULL COMMENT '目标ID（视频/评论ID）', 
    TYPE TINYINT NOT NULL COMMENT '类型：1视频 2评论', 
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', 
    deleted_at DATETIME COMMENT '软删除时间', 
    INDEX idx_user_id (user_id), 
    INDEX idx_target_id (target_id), 
    INDEX idx_user_target (user_id, target_id, TYPE), 
    INDEX idx_deleted_at (deleted_at) 
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='点赞表'; 

-- 关注关系表
CREATE TABLE FOLLOWS ( 
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '关注记录ID', 
    follower_id BIGINT NOT NULL COMMENT '关注者ID（谁发起的关注）', 
    following_id BIGINT NOT NULL COMMENT '被关注者ID（被关注的人）', 
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', 
    deleted_at DATETIME COMMENT '软删除时间（NULL表示未删除）', 
    UNIQUE KEY uk_follow (follower_id, following_id) COMMENT '防止重复关注', 
    INDEX idx_follower_id (follower_id) COMMENT '查询关注列表', 
    INDEX idx_following_id (following_id) COMMENT '查询粉丝列表', 
    INDEX idx_deleted_at (deleted_at), 
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE, 
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE 
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='关注关系表';
```

## 🛠️ 快速开始

### 1. 修改配置
编辑 `biz/dal/mysql/database.go`，确保数据库连接字符串与你的环境一致：
```go
db := "localhost"
port := "3306"
dataname := "video_web"
username := "root"
password := "your_password"
```

### 2. Docker 容器化部署
本项目推荐使用 Docker 进行部署：

**Dockerfile 示例**:
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY . .
RUN go mod download && go build -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
# 创建本地视频/头像存储目录
RUN mkdir -p uploads/videos uploads/avatars
EXPOSE 8888
CMD ["./server"]
```

## 📄 接口文档 (API)

完整的接口定义请参考项目中的 `idl/` 目录。
在线接口调试文档（Apifox）: `jsgyp49rqr.apifox.cn`

---
