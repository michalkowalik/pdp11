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

type InterruptStack []interrupts.Interrupt

func (i *InterruptStack) Push(interrupt interrupts.Interrupt) {
	*i = append(*i, interrupt)

	//if interrupt.Vector != interrupts.IntCLOCK {
	//	fmt.Printf("stack after push: %v\n", *i)
	//}

}

func (i *InterruptStack) Pop() (interrupts.Interrupt, error) {

	if len(*i) == 0 {
		fmt.Printf("popping from an empty interrupt stack \n")
		return interrupts.Interrupt{}, errors.New("interrupt stack is empty")
	}

	index := len(*i) - 1
	element := (*i)[index]
	*i = (*i)[:index] // truncate from stack

	//if element.Vector != uint16(interrupts.IntCLOCK) {
	//	fmt.Printf("popping %v from interrupt stack\n", element)
	//}
	return element, nil
}
