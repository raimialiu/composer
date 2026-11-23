package composer

func ThrowIfNull(params any) {
	if params == nil {
		panic("Argument cannot be nil")
	}
}
