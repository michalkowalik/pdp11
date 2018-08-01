# Unibus implementation notes
(For self remembering)

## 1. Memory mapped devices.
In this case -> if anything wants to read from a memory mapped device:
- Check address
- if fits:
    - send request to device to give it back (in case of reading)
    - send request to device to write (in case of write)
    - do not actually _read_ or _write_

### 1.1 Teletype write:
- system writes to memory address
- unibus identifies the address and checks if it's in a mapped IO range
- further down, unibus checks if it's a teletype address
- based on an identified address, `teletype.WriteTerm` is being called
- `teletype.WriteTerm` checks the address, and based on it, it either modifies
the control registers (`0560`, `0564`), or prints the data word (or more exactly it's lower byte), and sends back to Unibus the `TTYout` interrupt

### 1.2 Teletype read:
- Unibus identieds the address and checks if it's in a mapped IO range (within the `readIOPage`)
- if address matches the teletype range, the [`teletype.ReadTerm`](../teletype/teletype.go) is being called:
     - a control register status can be returned (`TKS` for `0560`, or `TPS` for `0566`) 
     - or the character in in the character buffer (`0564`)
     - unibus returns the word returned by `teletype.ReadTerm` as a result of `unibus.readIOPage`

### 1.3 Teletype continued.
In this particular implementation, it's the [`mmu`](../mmu/mmu.go) that takes care of finding out, if the initial memory address is within the range of memory mapped IO, and it's `mmu`'s job to call the Unibus. 

[`mmu`](../mmu/mmu.go) is returning the data it is getting back from unibus as a result for the read / write memory requests, that are located in the IO range.
No memory is written or being read during that process. it's all between `mmu` and `unibus`