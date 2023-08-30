package validator

import (
	"regexp"

	"github.com/h3ll0kitt1/avitotest/internal/models"
)

type Validator interface {
	UserId(user int64) bool
	Days(days int) bool
	PercentageRND(percentageRND int) bool
	SegmentSlug(slug string) bool
	Segments(segments []models.Segment) bool
}

type DefaultValidator struct {
	SegmentSlugExpr string
}

func New() *DefaultValidator {
	regularExpr := `^[a-zA-Z0-9_]*$`

	return &DefaultValidator{
		SegmentSlugExpr: regularExpr,
	}
}

func (v *DefaultValidator) UserId(user int64) bool {
	if user >= 1 {
		return true
	}
	return false
}

func (v *DefaultValidator) Days(days int) bool {
	if days >= 0 && days <= 5000 {
		return true
	}
	return false
}

func (v *DefaultValidator) PercentageRND(percentageRND int) bool {
	if percentageRND >= 0 && percentageRND <= 100 {
		return true
	}
	return false
}

func (v *DefaultValidator) SegmentSlug(slug string) bool {
	re := regexp.MustCompile(v.SegmentSlugExpr)
	return re.MatchString(slug)
}

func (v *DefaultValidator) Segments(segments []models.Segment) bool {
	re := regexp.MustCompile(v.SegmentSlugExpr)
	for _, segment := range segments {
		if !re.MatchString(segment.Slug) {
			return false
		}
		if segment.DaysTTL < 0 || segment.DaysTTL > 5001 {
			return false
		}
	}
	return true
}
