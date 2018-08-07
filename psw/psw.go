package psw

/**
Processor status word package
*/

// processor word layout. Values here are bits, not the
// powers of 2
const cFlag = 0
const vFlag = 1
const zFlag = 2
const nFlag = 3
const tFlag = 4

// PSW keeps processsor status word
type PSW uint16

// Get returns current processsor status word
func (psw *PSW) Get() uint16 {
	return uint16(*psw)
}

// GetMode returns 0 for user and 3 for kernel
func (psw *PSW) GetMode() uint16 {
	return uint16(*psw >> 14)
}

// C returns C flag:
func (psw *PSW) C() bool {
	return psw.getFlag(cFlag)
}

// SetC sets C flag
func (psw *PSW) SetC(status bool) {
	psw.setFlag(cFlag, status)
}

// V returns v flag
func (psw *PSW) V() bool {
	return psw.getFlag(vFlag)
}

// SetV sets processor V flag
func (psw *PSW) SetV(status bool) {
	psw.setFlag(vFlag, status)
}

// Z returns Z flag
func (psw *PSW) Z() bool {
	return psw.getFlag(zFlag)
}

// SetZ sets processor Z flag
func (psw *PSW) SetZ(status bool) {
	psw.setFlag(zFlag, status)
}

// N retruns N flag
func (psw *PSW) N() bool {
	return psw.getFlag(nFlag)
}

// SetN sets processor Z flag
func (psw *PSW) SetN(status bool) {
	psw.setFlag(nFlag, status)
}

// T returns T flag
func (psw *PSW) T() bool {
	return psw.getFlag(tFlag)
}

// SetT sets processor Z flag
func (psw *PSW) SetT(status bool) {
	psw.setFlag(tFlag, status)
}

// generic get flag function
func (psw *PSW) getFlag(flag uint) bool {
	return (*psw & (1 << flag)) > 0
}

// generic set flag function
func (psw *PSW) setFlag(flag uint, status bool) {
	if status == true {
		*psw |= (1 << flag)
	} else {
		*psw &^= (1 << flag)
	}
}
