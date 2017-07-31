package s3apt

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Method struct {
	input   *os.File
	output  *os.File
	session *Session
}

func NewMethod(input *os.File, output *os.File, session *Session) *Method {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}

	return &Method{
		input:   input,
		output:  output,
		session: session,
	}
}

func (m *Method) Process(message *Message) (err error) {
	if message.Capability == 600 {
		m.getS3(message)
	}
	return nil
}

func (m *Method) getS3(message *Message) {
	err := m.session.Get(message.Uri, message.Filename, m)
	if err != nil {
		errMsg := strings.Replace(err.Error(), "\n", "", -1)
		m.Send(&Message{
			Capability: 400,
			Message:    errMsg,
		})
	}
}

func (m *Method) Send(message *Message) {
	str, err := message.Output()
	if err != nil {
		m.Send(&Message{Capability: 401, Message: err.Error()})
		os.Exit(0)
	}
	fmt.Fprintf(m.output, "%s\n", str)
}

func (m *Method) SendExitError(err error) {
	m.Send(&Message{Capability: 401, Message: err.Error()})
	os.Exit(1)
}

func (m *Method) Capabilities(version string) {
	m.Send(&Message{
		Capability:     100,
		Version:        version,
		SingleInstance: true,
	})
}

func (m *Method) ParseCapability(line string) (capability int, err error) {
	r := regexp.MustCompile(`^(\d+)\s+(.+)\s*$`).FindAllStringSubmatch(line, -1)
	if len(r) != 1 || len(r[0]) != 3 {
		return 0, fmt.Errorf("Invalid capability input `%s`", line)
	}
	capability, _ = strconv.Atoi(r[0][1])
	return capability, nil
}

func (m *Method) ParseKV(line string) (key string, value string, err error) {
	r := regexp.MustCompile(`^([^:\s]+)\s*:\s*(.+)\s*$`).FindAllStringSubmatch(line, -1)
	if len(r) != 1 || len(r[0]) != 3 {
		return "", "", fmt.Errorf("Invalid 'key: value' input `%s`", line)
	}

	return r[0][1], r[0][2], nil
}
