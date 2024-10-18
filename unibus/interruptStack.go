package unibus

import (
	"errors"
	"fmt"
	"pdp/interrupts"
)

/*
 debug only:
 - push interrupt vector to the stack once it is being processed
 - pop from stack when rti comes

 - at no point an RTI should be attempted on an empty stack
*/

type InterruptStack []uint16

func (i *InterruptStack) Push(interrupt uint16) {
	*i = append(*i, interrupt)

	if interrupt != interrupts.INTClock {
		fmt.Printf("stack after push: %o\n", *i)
	}

}

func (i *InterruptStack) Pop() (uint16, error) {
	if len(*i) == 0 {
		fmt.Printf("popping from an empty interrupt stack \n")
		return 0, errors.New("interrupt stack is empty")
	}

	index := len(*i) - 1
	element := (*i)[index]
	*i = (*i)[:index] // truncate from stack
	if element != uint16(interrupts.INTClock) {
		fmt.Printf("popping %o from interrupt stack\n", element)
	}
	return element, nil
}
