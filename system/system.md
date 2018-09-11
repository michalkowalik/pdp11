# PDP-11 System module  

### Stable main loop

* excesive memory usage
* excesive cpu usage
* blocked (paused) goroutines

### Things to check:
1. Look for unibus events in the system.step() `for` loop
2. check how on earth is it solved in Cheney's port