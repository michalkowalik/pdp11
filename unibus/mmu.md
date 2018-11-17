# MMU 

## How does it work
PDP11/70 can be equipped with up to 4MB of RAM.
 As the memory addresses suppored by the cpu are 16bit long,
that means, there's only 64K of memory directly accessible to the program.
To circumvent it, the MMU (Memory Management Unit) has to be used.

#### Abbreviation used below:
* __MMR__ - Memory Management Register
* __PAR__ - Page Address Register
* __PDR__ - Page Description Register
* __PAF__ - Page Address Field
* __APF__ - Active Page Field
* __SR0__ - Status Register 0 
* __SR2__ - Status Register 2

on top of that, pdp 11/(44,70) are using Instruction and data memory pages
hence the I/D marker on the virtual address. 
That's also the reason for both MMUPar and MMUPDR

On a real PDP 11 the memory registers are located in thee uppermost 8K of RAM address space along with the Unibus I/O device registers.

#### MMR composition:
___(MMR registers are available on machines with 22 bit MMU (/70, /44))___
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
    * This actually raises the addressing question. is the address the  n-th word, or n-th byte in the memory?
    * The original 22 bit implementation was assuming the byte addressing, so whenever the word based addressing was used, the addresses, pointers etc. were updated by 2. Is it the same for the 18 bit? and does it make sense then, to keep the memory defined as 128k words?
* 16 Page Address Registers (PAR)
    * 8 of them are used in User mode, 8 in Kernel mode. The CPU mode is specified by bits 14 and 15 of the Processor Status Word (PSW)
* 16 Page Description Registers (PDR)
* 8 Active Page Registers (APR)

## PAR and PDR details
### PAR -> Page Address Register
PAR contains __Page Address Field__ - a 12 bit field, which specifies the starting address of the page as a __block__ number in physical memory.

Bits 12-15 are unused.

### PDR -> Page Description Register
```
15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 PDF
 X  ------------- PLF --------------  X   W   X   X  ED  - ACF -  X     
```
* X - unused
* ACF - Access Control Field:

    | ACF | Mode |
    |-----|------|
    | 00  | Non-residend, abort all accesses |
    | 01  | Read-only, abort on write attempt|
    | 10  | unused, abort all accesses|
    | 11  | read / write|
* ED - Expansion direction: Specifies whether the segment expands upward from __relative zero__ (ED=0), or downwards  toward the relative zero. Relative Zero is in this case  the __PAF__.
* W - indicates whether or not the page has been modified since the PSR (what is psr??) was loaded (W=1 is affirmative).
W bit is set to 0 on every PAR or PDR modification
* PLF - specifies number of blocks in the page. A page contains between 1 and 128 blocks of contiguos memory location. If the page expands downwards (ED=1), the page contains 128 - (page lenght in blocks).

## Status Registers (SR0 and SR2)
### SR0 (Status and error indicators)
```
| 15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |SR0
|ANR |ALE |AROV| X  | X  | X  | X | MM| X | MODE  | X |  PAGE NR  |EM |
```
* X   - Unused
* ANR - Abort: Non-resident
* ALE - Abort: Page Lenght Error
* AROV - Abort: Read-Only Access Violation
* MM - Maintenance Mode
* M - Mode (Kernel = 00, User = 11)
* PAGE NR - Page Number of a reference causing an error
* EM - Enable Management. When EM=1 all addresses are relocated by MMU

### SR2
Status Register 2 is loaded with 16 bit virtual address at the beginning of each instruction fetch. SR2 is read-only, can not be rewritten. SR2 is the Virtual Address Programm Counter


### 18 Bit Address calculation:
of a 16 bit virtual address:
* bits 15 to 13: `APF` (`Active Page Field`) - to determine which of the 8 APRs will be used to form the physical address
* bits 12 to 0: `Displacement Field` - this 13 bit field contains address relative to the beginning of the page. 13 bits allows 4k words of size.The Displacement Field is further divided into:
    * bytes 12 to 6: those 7 bits form the `Block Number` - the block number within the current page
    * bytes 5 to 0: those 6 bits form the `Displacement In Block (DIB)` - displacement in block referred to by the `Block Number`
* Remainder of the information required to form physical address:
    * lower 12 bits of the `Active Page Register (APR)` - called `PAF - Page Address Field` - specifies the starting address of the memory, which that APR rescribes. it's the ___block___ number in the memory, so if PAF=3 -> starting address = 96 (3 * 32 = 96)

To build a physical address:
1) determine APR
2) read PAF from APR
3) `PAF` + `Block Number` from displacement field form the starting address of a 32 word block
4) adding `DIB` forms final physical address.


### 18 Bit implementation details
#### 1) WriteMemoryWord
* WriteMemoryWord does uptdate the PSW, but there's no need to update separately CPU Mode (user / kernel), as the mode is itself derived from PSW state