# MMU 

## How does it work
PDP11/70 can be equipped with up to 4MB of RAM.
 As the memory addresses suppored by the cpu are 16bit long,
that means, there's only 64K of memory directly accessible to the program.
To circumvent it, the MMU (Memory Management Unit) has to be used.

#### Abbreviation used below:
* __MMR__ - Memory Management Register
* __PAR__ - Page Address Register
* __PAF__ - Page Address Field
* __APF__ - Active Page Field

on top of that, pdp 11/(44,70) are using Instruction and data memory pages
hence the I/D marker on the virtual address. 
That's also the reason for both MMUPar and MMUPDR

On a real PDP 11 the memory registers are located in thee uppermost 8K of RAM address space along with the Unibus I/O device registers.

#### MMR composition:
```
15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 MMR0
nonr leng read trap unus unus ena mnt cmp  -mode- i/d  --page--   enable
```


## TODOS
- extract mmu interface
- add an 18 bit MMU implementation. it seems  to be easier and smaller.
- also memory mapped IO are easier to define in 18 bit version
- How is actually MMU initialized? What is the original assignment of APRs and Page Addresses?


## 18 Bit MMU implementation:
* Easier, smaller than 22 bit. Max 256KB of RAM.
* Enough to run Unix

### 18 Bit MMU elements:
* Follows MMU implemenation of ___PDP-11/40___
* 128K Words array of `uint16` type as main physical memory
* 16 Page Address Registers (PAR)
    * 8 of them are used in User mode, 8 in Kernel mode. The CPU mode is specified by bits 14 and 15 of the Processor Status Word (PSW)
* 16 Page Description Registers (PDR)
* 8 Active Page Registers (APR)