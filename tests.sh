#!/bin/bash

for d in `ls -d */`; do go test pdp/$d; done
