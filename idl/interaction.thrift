namespace go interaction.example

include "common.thrift"

// ==================== 点赞相关 ====================

// 点赞操作请求
struct LikeRequest {
    1: i64 TargetID (api.form="target_id" api.json="target_id")    // 目标ID（视频/评论ID）
    2: i32 Type (api.form="type" api.json="type")                  // 类型：1视频 2评论
    3: bool Status (api.form="status" api.json="status")           // 点赞状态：true点赞 false取消点赞
}

// 点赞列表请求
struct LikeListRequest {
    1: i64 TargetID (api.form="target_id" api.json="target_id")    // 目标ID
    2: i32 Type (api.form="type" api.json="type")                  // 类型：1视频 2评论
    3: i32 Page (api.form="page" api.json="page")
    4: i32 PageSize (api.form="page_size" api.json="page_size")
}

// ==================== 评论相关 ====================

// 发表评论请求
struct CreateCommentRequest {
    1: i64 VideoID (api.form="video_id" api.json="video_id")       // 视频ID
    2: i64 ParentID (api.form="parent_id" api.json="parent_id")    // 父评论ID（0表示一级评论）
    3: string Content (api.form="content" api.json="content")      // 评论内容
}

// 评论列表请求
struct CommentListRequest {
    1: i64 VideoID (api.form="video_id" api.json="video_id")       // 视频ID
    2: i32 Page (api.form="page" api.json="page")
    3: i32 PageSize (api.form="page_size" api.json="page_size")
}

// 删除评论请求
struct DeleteCommentRequest {
    1: i64 CommentID (api.form="comment_id" api.json="comment_id") // 评论ID
}

// ==================== 服务定义 ====================

// 点赞服务（需要登录）
service LikeAuthService {
    // 点赞/取消点赞
    common.CommonResponse LikeAction(
        1: LikeRequest req
    ) (api.post="/api/likes");

    // 获取点赞列表
    common.CommonResponse GetLikeList(
        1: LikeListRequest req
    ) (api.get="/api/likes/list");
}

// 评论服务（需要登录）
service CommentAuthService {
    // 发表评论
    common.CommonResponse CreateComment(
        1: CreateCommentRequest req
    ) (api.post="/api/comments");

    // 删除评论
    common.CommonResponse DeleteComment(
        1: DeleteCommentRequest req
    ) (api.delete="/api/comments/:comment_id");
}

// 评论公共服务（无需登录）
service CommentPublicService {
    // 获取评论列表
    common.CommonResponse GetCommentList(
        1: CommentListRequest req
    ) (api.get="/api/videos/:video_id/comments");
}