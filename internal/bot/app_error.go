// SPDX-License-Identifier: GPL-2.0
package bot

import (
	"fmt"
	"net/http"
)

type AppError struct {
	w       http.ResponseWriter
	code    int
	err     error
	message string
}

func (re *AppError) Error() string {
	var msg string

	switch re.code {
	case 400:
		msg = fmt.Sprintf("Bad request. %s", re.message)
	case 404:
		msg = fmt.Sprintf("Not found. %s", re.message)
	case 500:
		msg = fmt.Sprintf("Internal server error. "+
			"Please notify admins. %s", re.message)
	default:
		msg = fmt.Sprintf("Unknown status Please notify admins. %s", re.message)
	}

	http.Error(re.w, msg, re.code)
	return fmt.Sprintf("%d: %s;\n",
		re.code, msg) + re.err.Error()
}
