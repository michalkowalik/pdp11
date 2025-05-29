package logger

import (
	"log"
	"os"
)

func New(path string) *log.Logger {
	if len(path) == 0 {
		return log.New(os.Stdout, "PDP", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		l := log.New(f, "PDP ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Printf("Initializing pdp11.log")
		return l
	}
}
