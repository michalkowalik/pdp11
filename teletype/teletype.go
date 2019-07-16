package teletype

// Teletype interface
type Teletype interface {
	Run() (err error)
	WriteTerm(address uint32, data uint16) (err error)
	ReadTerm(address uint32) (d uint16, err error)
	GetIncoming() chan Instruction
}
