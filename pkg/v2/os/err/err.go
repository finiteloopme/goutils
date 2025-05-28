package err

import "github.com/finiteloopme/goutils/pkg/log"

func IsError(err error) bool { return err != nil }

func PanicIfError(msg string, err error) {
	if err != nil {
		log.Warn(msg, err)
		panic(err)
	}
}

func WarnIfError(msg string, err error) {
	if err != nil {
		log.Warn(msg, err)
	}
}
