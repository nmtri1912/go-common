package errorutils

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewGrpcError(code codes.Code, message, reason, domain string, metadata map[string]string) error {
	st := status.New(code, message)
	nst, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	})
	if err != nil {
		return st.Err()
	}
	return nst.Err()
}
func ExtractReasonAndDomainFromError(err error, defaultDomain string) (codes.Code, string, string) {
	status, ok := status.FromError(err)
	if !ok {
		return codes.Internal, err.Error(), defaultDomain
	}
	if len(status.Details()) <= 0 {
		return status.Code(), err.Error(), defaultDomain
	}
	errInfo, ok := status.Details()[0].(*errdetails.ErrorInfo)
	if !ok {
		return status.Code(), err.Error(), defaultDomain
	}
	return status.Code(), errInfo.GetReason(), errInfo.GetDomain()
}
