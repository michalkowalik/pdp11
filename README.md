# README #

### What is this repository for? ###

* PDP11/40 emulator written in go
* It's still a very early stage - pretty much nothing works yet, and that's OK - I'll get there

### quick summary ###

- a PDP11/70 emulator like many others, but this time in Golang.
- The plan is to use a simplified Forth interpreter as a replacement for the PDP hardware console, to achieve the following:
  - halt
  - start
  - step
  - boot
  - show registers
  - modify registers
  - mount / unmount drives
  - ?

### What works already ###
Nothing really. 
CPU implementation is completed and more or less tested.
Teletype implementation is almost done.
Machine can boot, but that's all.
### How do I get set up? ###

* clone repo
* `go install pdp`

### Contribution guidelines ###

* feel free to ping me if you find the thing interesting

### Who do I talk to? ###

* me
