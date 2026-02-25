namespace go common.example

enum ErrorCode {
    SUCCESS = 0                // 成功
    REQUEST_ERROR              // 获取请求失败
    PARAM_ERROR = 1001         // 参数错误
    USER_NOT_LOGIN = 2001      // 用户未登录
    USER_EXIST = 2002          // 用户已存在
    USER_NOT_EXIST = 2003      // 用户不存在
    USER_PASSWORD_ERROR = 2004 // 用户密码错误
    VIDEO_NOT_EXIST = 3001     // 视频不存在
    VIDEO_FORMAT_ERROR = 3002  // 视频格式错误
    COMMENT_NOT_EXIST = 4001   // 评论不存在
    OPERATION_FORBIDDEN = 5001 // 操作禁止（如删他人评论）
    PROGRESS_ERROR = 6001      // 其他错误
}

struct CommonResponse {
    1: ErrorCode Code (api.json="code")
    2: string Message (api.json="message")
    3: binary Data (api.json="data")
}

struct Pagination {
    1: i32 Page (api.form="page" api.json="page" default='1')
    2: i32 PageSize (api.form="page_size" api.json="page_size" default='10')
}

struct IDRequest {
    1: string ID (api.form="id" api.json="id")
}

struct FileUploadRequest {
    1: string FileKey (api.form="file_key" api.json="file_key")
    2: string FileType (api.form="file_type" api.json="file_type")
}