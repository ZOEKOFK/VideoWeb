namespace go video.example

include "common.thrift"

struct UploadVideoRequest {
    1: string Title (api.form="title" api.json="title")
    2: string Description (api.form="description" api.json="description")
}

struct UserVideoListRequest {
    1: i32 Page (api.form="page" api.json="page")
    2: i32 PageSize (api.form="page_size" api.json="page_size")
}

struct SearchVideoRequest {
    1: string Keyword (api.form="keyword" api.json="keyword")
    2: i32 Page (api.form="page" api.json="page")
    3: i32 PageSize (api.form="page_size" api.json="page_size")
    4: string Sort (api.form="sort" api.json="sort")
}

struct HotVideoRequest {
    1: i32 Limit (api.form="limit" api.json="limit")
    2: string Type (api.form="type" api.json="type")
    3: i32 Page (api.form="page" api.json="page")
}

service VideoPublicService {
    common.CommonResponse SearchVideos(
        1: SearchVideoRequest req
    ) (api.get="/api/videos/search");

    common.CommonResponse GetHotVideos(
        1: HotVideoRequest req
    ) (api.get="/api/videos/hot");

    common.CommonResponse GetUserVideos(
        1: UserVideoListRequest req
    ) (api.get="/api/users/:user_id/videos");
}

service VideoAuthService {
    common.CommonResponse UploadVideo(
        1: UploadVideoRequest req
    ) (api.post="/api/videos");
}