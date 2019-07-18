package system

import (
	"pdp/unibus"
)

/*
	Minimal bootstrap code -> load to memory, start executing.
*/

const (
	// BOOTBASE is a base bootstrap address
	BOOTBASE = 02000
)

var bootcode = [...]uint16{
	0042113,        /* "KD" */
	0012706, 02000, /* MOV #boot_start, SP */
	0012700, 0000000, /* MOV #unit, R0        ; unit number */
	0010003,          /* MOV R0, R3 */
	0000303,          /* SWAB R3 */
	0006303,          /* ASL R3 */
	0006303,          /* ASL R3 */
	0006303,          /* ASL R3 */
	0006303,          /* ASL R3 */
	0006303,          /* ASL R3 */
	0012701, 0177412, /* MOV #RKDA, R1        ; csr */
	0010311,          /* MOV R3, (R1)         ; load da */
	0005041,          /* CLR -(R1)            ; clear ba */
	0012741, 0177000, /* MOV #-256.*2, -(R1)  ; load wc */
	0012741, 0000005, /* MOV #READ+GO, -(R1)  ; read & go */
	0005002,        /* CLR R2 */
	0005003,        /* CLR R3 */
	0012704, 02020, /* MOV #START+20, R4 */
	0005005, /* CLR R5 */
	// 000001,  /* WAIT (MK!) */
	0105711, /* TSTB (R1)  (wait for ready) */
	0100376, /* BPL .-2 */
	0105011, /* CLRB (R1) */
	0005007 /* CLR PC */}

// Boot loads bootstrap code and start emulation
func (sys *System) Boot() {
	memPointer := uint16(BOOTBASE)

	for _, c := range bootcode {
		sys.unibus.Mmu.WriteMemoryWord(memPointer, c)
		memPointer += 2
	}

	// set SP and PC to their starting address:
	sys.CPU.Registers[7] = BOOTBASE + 2

	// start execution
	// sys.console.WriteConsole("Booting..\n")
	if sys.CPU.State != unibus.CPURUN {
		sys.CPU.State = unibus.CPURUN
	}
	sys.Run()
}
