#!/bin/bash

for d in `ls -d */`; do go.exe test pdp/$d; done
