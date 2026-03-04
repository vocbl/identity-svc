package port

import (
	"context"

	app "github.com/vocbl/users-svc/internal/application/onboard"
	"github.com/vocbl/users-svc/internal/port/grpc/pb"
)

type OnboardHandler struct {
	service *app.OnboardService
}

func (h *OnboardHandler) Register(ctx context.Context, req *pb.RegistrationRequest) (*pb.RegistrationResult, error) {
	newUser := app.NewUser{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Username:  req.GetUsername(),
	}

	sessionID, err := h.service.CreateVerificationSession(ctx, newUser)
	if err != nil {
		return nil, ErrRegisterHandler.Wrap(err)
	}

	return &pb.RegistrationResult{
		SessionID: sessionID.String(),
	}, nil
}
