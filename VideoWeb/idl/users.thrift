namespace go user.example

include "common.thrift"

struct UserRegisterRequest {
    1: string Username (api.form="username" api.json="username")
    2: string Password (api.form="password" api.json="password")
    3: string Nickname (api.form="nickname" api.json="nickname")
}

struct UserLoginRequest {
    1: string Username (api.form="username" api.json="username")
    2: string Password (api.form="password" api.json="password")
    3: bool Remember (api.form="remember" api.json="remember")
}

service UserPublicService {
    common.CommonResponse Register(
        1: UserRegisterRequest req
    ) (api.post="/api/users");

    common.CommonResponse Login(
        1: UserLoginRequest req
    ) (api.post="/api/sessions");
}

service UserAuthService {
    common.CommonResponse GetUserInfo(
        1: common.IDRequest req
    ) (api.get="/api/users/:user_id");

    common.CommonResponse UploadAvatar(
    ) (api.put="/api/users/:user_id/avatar");
}