package cli

// Execute runs the perch root command (os.Args).
func Execute() error {
	return NewRootCmd().Execute()
}
