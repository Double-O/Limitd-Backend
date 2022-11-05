package utils

// controller&service
const InvalidGoogleTokenMsg = "validating google token failed, token: %+v, googleClientID: %+v"
const MismatchTokenAndLoginReqMsg = "mismatch in Claims and RequestBody in field: %+v, Claims : %+v, requestBody: %+v"
const InvalidThirdPartyIssuerMsg = "Invalid 3rd Party Issuer, From : %+v"
const CreateUserErrorMsg = "Error while creating user in db, err : %+v"
const FindUserByEmailErrorMsg = "Error while querying userByEmail in db, err : %+v"
const UserAlreadyExistMsg = "User already Exists"

// utils
const AccessTokenGenerationFailedMsg = "Failed to generate access token, err : %+v"
const RefreshTokenGenerationFailedMsg = "Failed to generate refresh token, err : %+v"
const FailedToParseAccessTokenMsg = "Failed to parse Access Token err : %+v"
const InvalidAccessTokenMsg = "Invalid Token"
const AccessTokenNotFoundMsg = "access token uuid not found"
