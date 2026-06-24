package request

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
)

func validateSiteID(siteID uint64) error {
	if siteID == 0 {
		return errors.New("SiteID is required")
	}

	return nil
}

func validateFilterValues(filter Filter) []error {
	errs := make([]error, 0)

	switch filter.Dimension.(type) {
	case dimensions.EventMetaKey:
		for _, values := range filter.Values {
			v, ok := values.(string)

			if !ok {
				errs = append(errs, fmt.Errorf("metadata key value must be a string"))
				break
			}

			if err := validateMetadataKey(v); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, f := range filter.Filter {
		errs = append(errs, validateFilterValues(f)...)
	}

	return errs
}

func validateMetadataKey(path string) error {
	if path == "" {
		return errors.New("metadata key path must not be empty")
	}

	parts := strings.Split(path, ".")

	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("metadata key path '%s' must not contain empty segments", path)
		}

		if !metaKeyValidKeySegment.MatchString(part) {
			return fmt.Errorf("metadata key path '%s' segment '%s' contains invalid characters: only a-z, A-Z, 0-9, _ and - are allowed", path, part)
		}
	}

	return nil
}
