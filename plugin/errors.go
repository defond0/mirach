package plugin

import (
	"gitlab.eng.cleardata.com/dash/mirach/util"
)

// Exception is a type that should be created and returned or used as an argument
// to panic when an error occurs while a plugin is running.
// In these cases, the caller of the plugin can choose to handle this error
// differently than unknown errors by reflecting on the type of the error.
type Exception struct {
	msg string
}

func (e Exception) Error() string {
	return e.msg
}

// ExceptionOrError returns an error of our custom type Error or the base error
// based on the errors presence in the list of known exceptions given.
func ExceptionOrError(err error, exceptions []string) error {
	if excString, _ := util.CheckExceptions(err, exceptions); excString != "" {
		return Exception{excString}
	}
	return err
}
