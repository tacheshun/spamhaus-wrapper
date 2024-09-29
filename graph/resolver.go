package graph

import (
	"context"
	"github.com/google/uuid"
	"spamhaus-wrapper/graph/generated"
	"spamhaus-wrapper/graph/model"
	"time"
)

type Resolver struct {
	IpDetails map[string]*model.IPDetails
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) UpdateIPDetails(ctx context.Context, ips []string) ([]*model.IPDetails, error) {
	var result []*model.IPDetails
	for _, ip := range ips {
		details := &model.IPDetails{
			UUID:         uuid.New().String(),
			IPAddress:    ip,
			ResponseCode: "200",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		r.IpDetails[ip] = details
		result = append(result, details)
	}
	return result, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetIPDetails(ctx context.Context, ip string) (*model.IPDetails, error) {
	if details, ok := r.IpDetails[ip]; ok {
		return details, nil
	}
	return nil, nil
}
