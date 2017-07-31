package main

import (
	"bufio"
	"github.com/sonofbytes/s3apt"
	"os"
)

var (
	VERSION = ""
	BRANCH  = ""
	COMMIT  = ""
)

func main() {
	message := &s3apt.Message{}
	session := s3apt.NewS3Apt()

	method := s3apt.NewMethod(nil, nil, session)
	method.Capabilities(VERSION)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if message != nil {
				method.Process(message)
			}
			message = &s3apt.Message{}
			continue
		}

		// First line of message should be capability
		if message.Capability == 0 {
			var err error
			message.Capability, err = method.ParseCapability(line)
			if err != nil {
				method.SendExitError(err)
			}
			continue
		}

		// Subsequent lines in key:value format
		k, v, err := method.ParseKV(line)
		if err != nil {
			method.SendExitError(err)
		}

		err = message.Set(k, v)
		if err != nil {
			method.SendExitError(err)
		}
	}
}
