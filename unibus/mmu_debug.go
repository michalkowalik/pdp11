package unibus

import (
	"fmt"
	"os"
)

// DumpMemory writes "size" words into file
func (m *MMU18) DumpMemory() error {
	file, err := os.Create("mem_dmp.txt")

	defer file.Close()
	for i := 0; i < 0760000/2; i++ {
		fmt.Fprintf(file, "%06o : %06o\n", i*2, m.unibus.Memory[i])
	}
	return err
}
