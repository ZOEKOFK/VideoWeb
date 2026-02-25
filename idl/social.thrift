    namespace go social.example

    include "common.thrift"

    struct FollowRequest {
        1: i64 UserID (api.form="user_id" api.json="user_id")
        2: bool Status (api.form="status" api.json="status")           // true关注 false取消关注
    }

    struct FollowListRequest {
        1: i32 Page (api.form="page" api.json="page")
        2: i32 PageSize (api.form="page_size" api.json="page_size")
    }

    struct FollowerListRequest {
        1: i32 Page (api.form="page" api.json="page")
        2: i32 PageSize (api.form="page_size" api.json="page_size")
    }

    struct FriendListRequest {
        1: i32 Page (api.form="page" api.json="page")
        2: i32 PageSize (api.form="page_size" api.json="page_size")
    }


    service FollowAuthService {
        // 关注/取消关注
        common.CommonResponse FollowAction(
            1: FollowRequest req
        ) (api.post="/api/follows");

        common.CommonResponse GetFriendList(
            1: FriendListRequest req
        ) (api.get="/api/users/friends");
    }

    service FollowPublicService {
        common.CommonResponse GetFollowList(
            1: FollowListRequest req
        ) (api.get="/api/users/:user_id/followings");

        common.CommonResponse GetFollowerList(
            1: FollowerListRequest req
        ) (api.get="/api/users/:user_id/followers");
    }