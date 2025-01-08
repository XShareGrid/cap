package errors

import (
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GRPCErr returns grpc error
func (e *UserError) GRPCErr(c codes.Code, lang language.Tag) error {
	s, trErr := e.TrError(lang)
	if trErr != nil {
		localLogger("failed to TrError: " + trErr.Error())
		return grpc.Errorf(c, e.Error())
	}
	return grpc.Errorf(c, s)
}
