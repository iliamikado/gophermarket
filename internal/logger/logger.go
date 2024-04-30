package logger

import "log"

func Log(msg interface{}) {
	log.Println(msg)
}