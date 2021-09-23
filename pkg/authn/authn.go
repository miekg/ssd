package authn

import "net/http"

// Identify returns nil if the request is found to be OK. The returned
// string is the user found while parsing the authentication info.
func Identify(r *http.Request) (string, error) {
	return "", nil
}
