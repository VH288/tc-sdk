package grpcerrors_test

import (
	"errors"
	"fmt"
	"testing"

	apperrors "github.com/VH288/tc-sdk/errors"
	"github.com/VH288/tc-sdk/grpcerrors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFromError(t *testing.T) {
	const message = "application message"
	tests := []struct {
		name    string
		err     error
		code    codes.Code
		message string
	}{
		{"not found", apperrors.New(apperrors.CodeNotFound, message), codes.NotFound, message},
		{"validation", apperrors.New(apperrors.CodeValidation, message), codes.InvalidArgument, message},
		{"bad request", apperrors.New(apperrors.CodeBadRequest, message), codes.InvalidArgument, message},
		{"conflict", apperrors.New(apperrors.CodeConflict, message), codes.AlreadyExists, message},
		{"unauthorized", apperrors.New(apperrors.CodeUnauthorized, message), codes.Unauthenticated, message},
		{"forbidden", apperrors.New(apperrors.CodeForbidden, message), codes.PermissionDenied, message},
		{"unavailable", apperrors.New(apperrors.CodeUnavailable, message), codes.Unavailable, message},
		{"timeout", apperrors.New(apperrors.CodeTimeout, message), codes.DeadlineExceeded, message},
		{"internal", apperrors.New(apperrors.CodeInternal, message), codes.Internal, message},
		{"unknown application code", apperrors.New(apperrors.Code("CUSTOM"), message), codes.Internal, message},
		{"wrapped application error", fmt.Errorf("outer: %w", apperrors.New(apperrors.CodeNotFound, message)), codes.NotFound, message},
		{"ordinary error", errors.New("raw unexpected failure"), codes.Internal, "internal error"},
	}

	require.NoError(t, grpcerrors.FromError(nil))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcerrors.FromError(tt.err)
			s, ok := status.FromError(got)
			require.True(t, ok)
			require.Equal(t, tt.code, s.Code())
			require.Equal(t, tt.message, s.Message())
		})
	}
}
