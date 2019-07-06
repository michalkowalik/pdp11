#!/bin/bash

MODULES="console interrupts mmu psw system teletype unibus"

for M in $MODULES; do go test pdp/$M; done

# for d in `ls -d */`; do go test pdp/$d; done
