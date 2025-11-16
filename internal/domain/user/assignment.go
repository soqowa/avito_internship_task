package user

import (
	"context"

	"github.com/user/reviewer-svc/internal/domain"
)


type AssignmentStrategy interface {
	ChooseInitialReviewers(ctx context.Context, candidates []User, max int) ([]User, error)
	ChooseReassignment(ctx context.Context, oldReviewer User, candidates []User) (User, error)
}

type RandomAssignmentStrategy struct {
	rand domain.Rand
}

func NewRandomAssignmentStrategy(r domain.Rand) *RandomAssignmentStrategy {
	return &RandomAssignmentStrategy{rand: r}
}



func (s *RandomAssignmentStrategy) ChooseInitialReviewers(_ context.Context, candidates []User, max int) ([]User, error) {
	if max <= 0 || len(candidates) == 0 {
		return nil, nil
	}
	if len(candidates) <= max {
		return candidates, nil
	}

	idx := make([]int, len(candidates))
	for i := range idx {
		idx[i] = i
	}

	limit := max
	for i := 0; i < limit; i++ {
		j := i + s.rand.Intn(len(candidates)-i)
		idx[i], idx[j] = idx[j], idx[i]
	}

	res := make([]User, limit)
	for i := 0; i < limit; i++ {
		res[i] = candidates[idx[i]]
	}
	return res, nil
}


func (s *RandomAssignmentStrategy) ChooseReassignment(_ context.Context, oldReviewer User, candidates []User) (User, error) {
	if len(candidates) == 0 {
		return User{}, domain.ErrNoCandidate
	}
	_ = oldReviewer
	idx := s.rand.Intn(len(candidates))
	return candidates[idx], nil
}

