package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/62teknologi/62dolphin/app/tokens"
	dutils "github.com/62teknologi/62dolphin/app/utils"
	"google.golang.org/api/option"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/adapters"
	"github.com/62teknologi/62dolphin/app/config"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	adapter, err := adapters.GetAdapter(ctx.Param("adapter"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("success", err.Error(), nil))
	}

	adapter = adapter.Init()
	ctx.Redirect(http.StatusTemporaryRedirect, adapter.GenerateLoginURL())
}

func Callback(ctx *gin.Context) {
	adapterName := ctx.Param("adapter")

	if strings.Contains(ctx.FullPath(), "/auth/sign-in") {
		adapterName = "local"
	}

	adapter, err := adapters.GetAdapter(adapterName)
	if err != nil {
		ctx.JSON(
			http.StatusNotFound, utils.ResponseData(
				"error",
				err.Error(),
				nil,
			),
		)
	}

	adapter = adapter.Init()
	adapter.Callback(ctx)
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
}

func Verify(ctx *gin.Context) {
	var req OAuthToken

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	adapterName := ctx.Param("adapter")

	adapter, err := adapters.GetAdapter(adapterName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("success", err.Error(), nil))
	}

	adapter = adapter.Init()

	// Initialize the Firebase app.
	opt := option.WithCredentialsFile(config.Data.FirebaseConfigPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Obtain a client for the Firebase Auth service.
	authClient, err := app.Auth(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	firebaseProfile, err := authClient.VerifyIDToken(ctx, req.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	token, err := adapter.Verify(ctx, firebaseProfile.Claims["email"].(string), firebaseProfile.Claims["user_id"].(string))
	if err != nil {
		if err.Error() == "user not found" {
			//
			if err := utils.DB.Table("users").Create(&map[string]any{
				"email":             firebaseProfile.Claims["email"],
				adapterName + "_id": firebaseProfile.Claims["user_id"],
				"password":          "$2a$10$P9wjSPl0lcrJzSQucqi8OOdrjNVj.jgAFn7vYf6gcpoXwfRgXVHRG",
				"created_at":        time.Now(),
			}).Error; err != nil {
				if duplicateError := utils.DuplicateError(err); duplicateError != nil {
					authorizationHeader := ctx.GetHeader("authorization")
					fields := strings.Fields(authorizationHeader)
					var accessToken string
					if len(fields) > 1 {
						accessToken = fields[1]
					}

					if accessToken == "" {
						ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", "linking account need authorized user", nil))
						return
					}

					payload, _ := tokenMaker.VerifyToken(accessToken)

					var user map[string]any
					utils.DB.Table("users").Where("id = ?", payload.UserId).Take(&user)

					updatedOtp := utils.DB.Table("users").
						Where("email = ?", user["email"]).Updates(map[string]any{
						adapterName + "_id": firebaseProfile.Claims["user_id"],
					})
					if updatedOtp.Error != nil {
						if duplicateError := utils.DuplicateError(updatedOtp.Error); duplicateError != nil {
							ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "user id "+firebaseProfile.Claims["user_id"].(string)+" already linked to another account", nil))
							return
						}
						ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", updatedOtp.Error.Error(), nil))
						return
					}

					ctx.JSON(http.StatusOK, utils.ResponseData("success", fmt.Sprintf("success linking %s account", adapterName), nil))
					return
				}

				ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
				return
			}

			var user map[string]any
			utils.DB.Table("users").Where(adapterName+"_id = ?", firebaseProfile.Claims["user_id"]).Take(&user)

			ctx.JSON(http.StatusOK, utils.ResponseData("success", fmt.Sprintf("success create and linking %s account", adapterName), map[string]any{
				"id": user["id"],
			}))
			return
		}

		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "success verify user", token))
}

func ForgotPassword(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/auth/forgot_password.json")
	var userInput map[string]any
	if err = ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, userInput)
	utils.MapNullValuesRemover(input)

	// Check if user exist in db
	var user map[string]any
	utils.DB.Table("users").Where(input["method"].(string)+" = ?", input["receiver"]).Take(&user)

	if user["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid user", nil))
		return
	}

	otpCode, _ := dutils.GenerateOTP(8)
	otpParams := map[string]any{
		"type":       "email",
		"code":       otpCode,
		"receiver":   input["receiver"],
		"expires_at": time.Now().Local().Add(config.Data.ForgotPasswordExpireDuration),
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	hashedOtp, _ := dutils.HashPassword(otpCode)
	otpParams["code"] = hashedOtp

	createOtp := utils.DB.Table("otps").Create(otpParams)
	if createOtp.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createOtp.Error.Error(), nil))
		return
	}
	resetToken := utils.Encode(fmt.Sprintf("%v#%v#%v", input["method"], otpCode, input["receiver"]))

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "send forgot password otp successfully", map[string]any{
		"reset_token": resetToken,
	}))
}

type resetPasswordParams struct {
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func ResetPassword(ctx *gin.Context) {
	var req resetPasswordParams
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "new password is not the same as confirm password", nil))
		return
	}

	// decodePasswordToken should return <otp type>#<otp code>#<otp content>
	decodePasswordToken, err := utils.Decode(ctx.Param("token"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}
	splitToken := strings.Split(decodePasswordToken, "#")

	var otp map[string]any
	err = dutils.VerifyOTP(otp, splitToken[0], splitToken[2], splitToken[1])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Hashing Password
	hashedPassword, err := dutils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
		return
	}

	var user map[string]any
	utils.DB.Table("users").Where(splitToken[0]+" = ?", splitToken[2]).Update("password", hashedPassword)
	fmt.Println(user)

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "reset password successfully", nil))
}
