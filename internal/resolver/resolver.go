package resolver

import (
	"spamhaus-wrapper/internal/repository"
)

type Resolver struct {
	IPDetailsRepo *repository.IPDetailsRepository
}
