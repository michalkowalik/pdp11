# What do we know by now
- pushing enter does not cause the trap to occur.
  The keystroke is being acknowledged, the interrupt 060 (key entered) is sent and processed
  The Interrupt 064 (print on terminal) is processed & an empty line is printed.
  The prompt is printed, and system waits for next instruction. 
- all that, allows me to believe, that the basic interrupt processing works just fine.
- entering anything on the prompt ends up with the following:
  ```
  PDP 2025/05/25 10:42:30 simple.go:151: Adding char 10
  PDP 2025/05/25 10:42:30 system.go:160: processing interrupt with the vector 060
  PDP 2025/05/25 10:42:30 simple.go:92: Sending TTY interrupt 64
  PDP 2025/05/25 10:42:30 system.go:160: processing interrupt with the vector 064
  PDP 2025/05/25 10:42:30 simple.go:92: Sending TTY interrupt 64
  PDP 2025/05/25 10:42:30 system.go:160: processing interrupt with the vector 064
  PDP 2025/05/25 10:42:30 cpu.go:363: Switching CPU from 0 to 3 mode
  PDP 2025/05/25 10:42:30 instructions.go:501: calling rti
  PDP 2025/05/25 10:42:30 instructions.go:509: interrupt return in user mode
  PDP 2025/05/25 10:42:30 cpu.go:250: ERROR: Invalid instruction: 177770
  ```
  - keystrokes are received and processed.
  - interrupt 064 triggered an processed (`LS` is printed on the screen)
  - CPU enters user mode
  - an RTI is being called from the user mode
    - on itself, this can happen, but for the trap processing only (?)
    - the RTI in user mode has a wrong, but very, very constant vector of `177770`

# Open questions
- Why is RTT being called all the time for the interrupt return? 
  I get, they're almost the same, but why not RTI?
