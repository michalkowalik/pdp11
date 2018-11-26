# RK11 disk controller
* RK11 - disk controller
* RK05 - disk cardridge

## RK11 Registers:

### Drive Status Register (RKDS)
* unibus address: 0777400
* bits:

    |bits|designation|description|
    |----|-----------|-----------|
    |0 - 3 | Sector Counter (SC)|current sector address of the seclected drive|
    | 4 | SC = SA | Sector counter = Sector Address. Indicates that disk heads are positioned over the disk addr. currently kept in Sector Address|
    | 5 | Write Protect Status (WPS) | 1  if disk in write protect mode|
    | 6 | Read/Write/Seek Ready (W/R/S RDY)|Drive mechanism is not in motion, drive ready to receive next order|
    | 7 | Drive Ready (DRY) | Drive ready to work and not in DRU condition|
    | 8 | Sector Counter OK (SOK) | SC on selected drive not in process of changing, and can be examined. if not set, second attempt should be made|
    | 9 | Seek Incomplete (SIN) | Seek function can not be completed. Cleared by reset|
    | 10| Drive Unsafe (DRU) | Drive unable to complete operation. Reset may be required |
    | 11| RK05 disk online | always set, to identify inserted disk as rk05|
    | 12| Drive Power Low (DPL)| power loss sensed. reset may be required|
    | 13 - 15| Identification of drive (ID) | if interrupt occures, those bits keep the binary representation of logical drive number, that caused the interrupt|

        

### Error Register (RKER)
* unibus address: 0777402
* bits:

### Control Status Register (RKCS)
* unibus address: 0777404
* bits:

    |bits|designation|description|
    |----|-----------|-----------|
    | 0  | GO | Causes the control to carry the function of ID kept in bits 1-3. Remains set until control begins to respond to GO|
    | 7  | Control Ready (RDY - write only)| Control ready to perform a function. Set by init, cleared by GO|

### Word Count Register (RKWC)
* unibus address: 0777406

### Current Bus Address Register (RKBA)
* unibus address: 077410
* bits:
    * 0: unused:
    * 1-15: ???

### Disk Address Register (RKDA)    