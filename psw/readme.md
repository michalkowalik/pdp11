# PSW: Processor status word 
Probably it's a bit of a overdoing to export it to a  separate module, 
but on the other hand, it should be easier to manage, and if it's too slow,
refactoring is always an option

PSW is loaded by MMU to the location 0777776
lowest bits contain CPU flags, upper bits CPU mode

## PSW bits one by one:

|bit|meaning|
|---|-------|
| 0 | C |
| 1 | V |
| 2 | Z |
| 3 | N |
| 4 | T |
| 5 - 7 | Priority |
| 8 | CIS instruction suspension |
| 9 - 10| Reserved|
|11 | General Register set |
| 12 - 13 | Previous mode |
| 14 - 15 | Current Mode |
