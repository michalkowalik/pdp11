package console

// Console interface to allow easy debugging
// gui less operations
type Console interface {
	WriteConsole(msg string) (err error)
}
