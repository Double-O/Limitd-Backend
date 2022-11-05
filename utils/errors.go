package utils

// controller
const InvalidGoogleTokenMsg = "validating google token failed, token: %+v, googleClientID: %+v"
const MismatchTokenAndLoginReqMsg = "mismatch in Claims and RequestBody in field: %+v, Claims field value : %+v, requestBody field value: %+v"
const InvalidThirdPartyIssuerMsg = "Invalid 3rd Party Issuer, From : %+v"
const UserNotFoundInCtxMsg = "User not found in Context"
const AccessTokenUUIDNotFoundInCtxMsg = "Access Token UUID Not Found in Context"

// service
const CreateUserErrorMsg = "Error while creating user in db, err : %+v"
const FindUserByEmailErrorMsg = "Error while querying userByEmail in db, err : %+v"
const FindUserByUuidErrorMsg = "Error while querying userByUuid in db, err : %+v"

// utils
const AccessTokenGenerationFailedMsg = "Failed to generate access token, err : %+v"
const RefreshTokenGenerationFailedMsg = "Failed to generate refresh token, err : %+v"
const FailedToParseTokenMsg = "Failed to parse and verify %+v Token err : %+v"

const InvalidTypeOfTokenCallMsg = "Invalid Type of Token Call"
const InvalidTokenMsg = "Invalid %+v token"
const TokenUUIDNotFoundMsg = "%+v token uuid not found in redis"
const RefreshTokenNotFoundInCookieMsg = "Refresh Token not found in cookie"
