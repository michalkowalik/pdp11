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

// KernelMode - processor mode
const KernelMode = 0

// UserMode - processor mode
const UserMode = 3

// PSW keeps processor status word
type PSW uint16

// Get returns current processor status word
func (psw *PSW) Get() uint16 {
	return uint16(*psw)
}

// Set PSW value
func (psw *PSW) Set(p uint16) {
	*psw = PSW(p)
}

// Priority - current cpu priority
func (psw *PSW) Priority() uint16 {
	return uint16((*psw >> 5) & 7)
}

// GetMode returns 3 for user and 0 for kernel
func (psw *PSW) GetMode() uint16 {
	return uint16(*psw >> 14)
}

func (psw *PSW) IsUserMode() bool {
	return psw.GetMode() == 3
}

// GetPreviousMode returns previous system mode: 3 for user, 0 for kernel
func (psw *PSW) GetPreviousMode() uint16 {
	return uint16((*psw >> 12) & 03)
}

// SwitchMode sets CPU into user or kernel mode and saves previous mode to
// psw previous mode field (bits 12, 13)
// short reminder: 00 means kernel, b11 means user
// switch mode should also switch between user and kernel stacks!
func (psw *PSW) SwitchMode(m uint16) {
	currentMode := psw.GetMode()

	*psw &= 07777
	if m > 0 {
		*psw |= (1 << 15) | (1 << 14)
	}
	if currentMode > 0 {
		*psw |= (1 << 12) | (1 << 13)
	}
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

// N returns N flag
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

// GetFlags returns set flags
func (psw *PSW) GetFlags() string {
	var flags string
	if psw.GetPreviousMode() > 0 {
		flags = "u"
	} else {
		flags = "k"
	}
	if psw.GetMode() > 0 {
		flags += "U"
	} else {
		flags += "K"
	}

	if psw.N() {
		flags += "N"
	} else {
		flags += " "
	}
	if psw.Z() {
		flags += "Z"
	} else {
		flags += " "
	}
	if psw.V() {
		flags += "V"
	} else {
		flags += " "
	}
	if psw.C() {
		flags += "C"
	} else {
		flags += " "
	}
	return "[" + flags + "]"
}
