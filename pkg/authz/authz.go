package authz

import (
	"bufio"
	"os"
)

// IsAllowed checks is user is allowed to use systemd. The userfile is parsed on every
// request as it's assumed to be small enough.
func IsAllowed(user string, userfile string) (bool, error) {
	f, err := os.Open(userfile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() == user {
			return true, nil
		}
	}
	return false, scanner.Err()
}
