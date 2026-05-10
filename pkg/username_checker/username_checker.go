package usernamechecker

func Check(userName string) bool {

	if len(userName) < 3 || len(userName) > 40 {

		return false
	}

	return true
}
