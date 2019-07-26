package unibus

import (
	"fmt"
	"os"
)

// DumpMemory writes "size" words into file
func (m *MMU18Bit) DumpMemory() error {
	file, err := os.Create("mem_dmp.txt")

	defer file.Close()
	for i := range m.Memory {
		fmt.Fprintf(file, "%06o : %06o\n", i*2, m.Memory[i])
	}
	return err
}
