package pirsch

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
