/*
Package grpc provides user grpc server
*/
package grpc

import (
	"context"
	"fmt"

	"github.com/vardius/go-api-boilerplate/pkg/common/domain"
	"github.com/vardius/go-api-boilerplate/pkg/common/jwt"
	"github.com/vardius/go-api-boilerplate/pkg/user/application"
	"github.com/vardius/go-api-boilerplate/pkg/user/domain/user"
	"github.com/vardius/go-api-boilerplate/pkg/user/interfaces/proto"
)

func registerCommandHandlers(cb domain.CommandBus, es domain.EventStore, eb domain.EventBus, j jwt.Jwt) {
	cb.Subscribe(RegisterUserWithEmail, application.OnRegisterUserWithEmail(es, eb, j))
	cb.Subscribe(RegisterUserWithGoogle, application.OnRegisterUserWithGoogle(es, eb))
	cb.Subscribe(RegisterUserWithFacebook, application.OnRegisterUserWithFacebook(es, eb))
	cb.Subscribe(ChangeUserEmailAddress, application.OnChangeUserEmailAddress(es, eb))
}

func registerEventHandlers(eb domain.EventBus) {
	eb.Subscribe(fmt.Sprintf("%T", &user.WasRegisteredWithEmail{}), application.WhenUserWasRegisteredWithEmail)
	eb.Subscribe(fmt.Sprintf("%T", &user.WasRegisteredWithGoogle{}), application.WhenUserWasRegisteredWithGoogle)
	eb.Subscribe(fmt.Sprintf("%T", &user.WasRegisteredWithFacebook{}), application.WhenUserWasRegisteredWithFacebook)
	eb.Subscribe(fmt.Sprintf("%T", &user.EmailAddressWasChanged{}), application.WhenUserEmailAddressWasChanged)
}

type userServer struct {
	commandBus domain.CommandBus
	eventBus   domain.EventBus
	eventStore domain.EventStore
	jwt        jwt.Jwt
}

// DispatchCommand implements proto.UserServer interface
func (s *userServer) DispatchCommand(ctx context.Context, cmd *proto.DispatchCommandRequest) (*proto.DispatchCommandResponse, error) {
	out := make(chan error)
	defer close(out)

	go func() {
		s.commandBus.Publish(ctx, cmd.GetName(), cmd.GetPayload(), out)
	}()

	select {
	case <-ctx.Done():
		return new(proto.DispatchCommandResponse), ctx.Err()
	case err := <-out:
		return new(proto.DispatchCommandResponse), err
	}
}

// New returns new user server object
func New(cb domain.CommandBus, eb domain.EventBus, es domain.EventStore, j jwt.Jwt) proto.UserServer {
	s := &userServer{cb, eb, es, j}

	registerCommandHandlers(cb, es, eb, j)
	registerEventHandlers(eb)

	return s
}
