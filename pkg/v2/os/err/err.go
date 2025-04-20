package err

func IsError(err error) bool { return err != nil }

func ExitIfError(err error) {
	if err != nil {
		panic(err)
	}
}
