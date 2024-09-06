package client

// setServiceOptionsDefault sets the default values for ServiceOptions.
func setServiceOptionsDefault(options *ServiceOptions) {
	if options.Common != nil {
		options.Common.Complete()
	}
	if options.ConnectorCreator == nil {
		options.ConnectorCreator = NewConnector
	}
}

type cancelErr struct {
	Err error
}

func (e cancelErr) Error() string {
	return e.Err.Error()
}
