package devlog

import (
	"fmt"
)

func Log(usageContext, msg string, args ...interface{}) {
	if isDevelop() {
		line := fmt.Sprintf(msg)
		if len(args) > 0 {
			line = fmt.Sprintf(msg, args...)
		}
		fmt.Printf("(%s) %s\n", usageContext, line)
	}
}
func Logger(usageContext string) func (string, ...interface{}) {
	return func(msg string, args ...interface{}) {
		Log(usageContext, msg, args...)
	}
}

// // format all errors consistently, and provide context for the error using the string `msg`
// func Eout(err error, msg string, args ...interface{}) error {
// 	if err != nil {
// 		// received an invocation of e.g. format:
// 		// Eout(err, "reading data for %q and %q", "database item", "weird user")
// 		if len(args) > 0 {
// 			return fmt.Errorf("%s (%w)", fmt.Sprintf(msg, args...), err)
// 		}
// 		return fmt.Errorf("%s (%w)", msg, err)
// 	}
// 	return nil
// }
