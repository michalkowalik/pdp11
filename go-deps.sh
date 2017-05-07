#!/bin/sh

echo "installing go dependencies.."
PACKAGES="github.com/jroimartin/gocui"


## missing gorilla packages?

for P in $PACKAGES 
do
    echo "installing $P.."
    go get -u -v $P
done
