// Package grpcerrors converts SDK application errors to gRPC status errors.
package grpcerrors

import (
	apperrors "github.com/VH288/tc-sdk/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError converts an error to a gRPC status, hiding non-application errors.
func FromError(err error) error {
	if err == nil {
		return nil
	}

	appErr, ok := apperrors.As(err)
	if !ok {
		return status.Error(codes.Internal, "internal error")
	}

	switch appErr.Code {
	case apperrors.CodeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case apperrors.CodeValidation, apperrors.CodeBadRequest:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case apperrors.CodeConflict:
		return status.Error(codes.AlreadyExists, appErr.Message)
	case apperrors.CodeUnauthorized:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperrors.CodeForbidden:
		return status.Error(codes.PermissionDenied, appErr.Message)
	case apperrors.CodeUnavailable:
		return status.Error(codes.Unavailable, appErr.Message)
	case apperrors.CodeTimeout:
		return status.Error(codes.DeadlineExceeded, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}
