package emailformatechecker

import "regexp"

var emailRx = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func Check(email string) bool {

	return emailRx.MatchString(email)
}
