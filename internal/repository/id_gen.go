package repository

import (
	"io"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	uuidgen "github.com/google/uuid"
)

// GenerateUUID returns a random UUID
func GenerateUUID(rand io.Reader) (ref.UUID, error) {
	uuidgen.SetRand(rand)

	uuid, err := uuidgen.NewRandom()
	if err != nil {
		return "", domain.WrapErrorf(err, domain.ErrorCodeUnknown, "Could not generate UUID")
	}

	return ref.UUID(uuid.String()), nil
}
