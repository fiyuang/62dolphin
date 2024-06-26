package middlewares

import (
	"errors"
	"fmt"
	"github.com/62teknologi/62dolphin/app/config"
	"net/http"
	"strings"

	dutils "github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/tokens"
	"github.com/62teknologi/62dolphin/app/utils"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(tokenMaker tokens.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

			return
		}

		fields := strings.Fields(authorizationHeader)

		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

			return
		}

		authorizationType := strings.ToLower(fields[0])

		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

			return
		}

		accessToken := fields[1]

		// check blocked access token
		if config.Data.TokenDestroy == true || config.Data.SimultaneousSession == false {
			var token map[string]any
			dutils.DB.Table("tokens").Where("access_token = ?", accessToken).Take(&token)
			if token["is_blocked"].(int8) == 1 {
				err := errors.New("token unauthorized")
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))
				return
			}
		}

		payload, err := tokenMaker.VerifyToken(accessToken)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
