# PSW: Processor status word 
Probably it's a bit of a overdoing to export it to a  separate module, 
but on the other hand, it should be easier to manage, and if it's too slow,
refactoring is always an option

PSW is loaded by MMU to the location 0777776
lowest bits contain CPU flags, upper bits CPU mode