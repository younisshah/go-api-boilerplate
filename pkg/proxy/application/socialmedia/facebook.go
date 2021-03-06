package socialmedia

import (
	"net/http"

	"github.com/vardius/go-api-boilerplate/pkg/common/http/response"
	"github.com/vardius/go-api-boilerplate/pkg/common/jwt"
	"github.com/vardius/go-api-boilerplate/pkg/common/security/identity"
	user_grpc "github.com/vardius/go-api-boilerplate/pkg/user/interfaces/grpc"
	user_proto "github.com/vardius/go-api-boilerplate/pkg/user/interfaces/proto"
)

type facebook struct {
	client user_proto.UserClient
	jwt    jwt.Jwt
}

func (f *facebook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessToken := r.FormValue("accessToken")
	data, e := getProfile(accessToken, "https://graph.facebook.com/me")
	if e != nil {
		response.WithError(r.Context(), response.HTTPError{
			Code:    http.StatusBadRequest,
			Error:   e,
			Message: "Invalid access token",
		})
		return
	}

	identity := &identity.Identity{}
	identity.FromFacebookData(data)

	token, e := f.jwt.Encode(identity)
	if e != nil {
		response.WithError(r.Context(), response.HTTPError{
			Code:    http.StatusInternalServerError,
			Error:   e,
			Message: "Generate token failure",
		})
		return
	}

	payload := &commandPayload{token, data}
	_, e = f.client.DispatchCommand(r.Context(), &user_proto.DispatchCommandRequest{
		Name:    user_grpc.RegisterUserWithFacebook,
		Payload: payload.toJSON(),
	})

	if e != nil {
		response.WithError(r.Context(), response.HTTPError{
			Code:    http.StatusBadRequest,
			Error:   e,
			Message: "Invalid request",
		})
		return
	}

	response.WithPayload(r.Context(), &responsePayload{token, identity})
	return
}

// NewFacebook creates facebook auth handler
func NewFacebook(c user_proto.UserClient, j jwt.Jwt) http.Handler {
	return &facebook{c, j}
}
