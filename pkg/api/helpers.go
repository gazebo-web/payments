package api

import "net/url"

// validateURL validates if a raw URL string is well-formed or not.
func validateURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	if u.Scheme == "" && u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
