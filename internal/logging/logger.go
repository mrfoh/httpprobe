package logging

func NewLogger(options *LoggerOptions) (Logger, error) {
	err := options.Validate()
	if err != nil {
		return nil, err
	}

	return NewZapLogger(options)
}
