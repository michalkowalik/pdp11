# PDP 11 CPU module

## Addressing

Instruction operands are six bits in length - three bits for the mode and three for the register. 

The 17th I/D bit in the resulting virtual address represents whether the reference is to **Instruction space** or **Data space** - which depends on combination of the mode and whether the register is the Program Counter (register 7).

### Addressing Modes:
The eight modes are:-
|Mode|Mnemonic|Description|
|----|--------|-----------|
|0|R|no valid virtual address|
|1|(R)|operand from I/D depending if R = 7|
|2|(R)+|operand from I/D depending if R = 7|
|3|@(R)+|address from I/D depending if R = 7 and operand from D space|
|4|-(R)|operand from I/D depending if R = 7|
|5|@-(R)|address from I/D depending if R = 7 and operand from D space|
|6|x(R)|x from I space but operand from D space|
|7|@x(R)|x from I space but address and operand from D space|

Stack limit checks are implemented for modes 1, 2, 4 & 6 (!)

#### **!!**
Also need to keep CPU.MMR1 updated as this stores which registers have been incremented and decremented so that the OS can reset and restart an instruction if a page fault occurs.
