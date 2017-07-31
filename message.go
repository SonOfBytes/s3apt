package s3apt

import (
	"errors"
	"fmt"
	"gopkg.in/oleiade/reflections.v1"
	"os"
	"reflect"
	"strconv"
	"time"
)

var messageCodes = map[int]string{
	100: "Capabilities",
	102: "Status",
	200: "URI Start",
	201: "URI Done",
	400: "URI Failure",
	401: "General Failure",
	600: "URI Acquire",
	601: "Configuration",
}

type Message struct {
	Capability        int
	capabilityMessage string
	FailIgnore        bool       `output:"Fail-Ignore"`
	Filename          string     `output:"Filename"`
	IndexFile         bool       `output:"Index-File"`
	LastModified      *time.Time `output:"Last-Modified"`
	MD5Hash           string     `output:"MD5-Hash"`
	MD5SumHash        string     `output:"MD5Sum-Hash"`
	SHA256Hash        string     `output:"SHA256-Hash"`
	SHA512Hash        string     `output:"SHA512-Hash"`
	Message           string     `output:"Message"`
	SingleInstance    bool       `output:"Single-Instance"`
	Size              int        `output:"Size"`
	Uri               string     `output:"URI"`
	Version           string     `output:"Version"`
}

func (m *Message) Set(key string, value string) (err error) {
	structTags, _ := reflections.Tags(m, "output")
	for k, o := range structTags {
		if o == key {
			key = k
			break
		}
	}

	if ok, err := reflections.HasField(m, key); err != nil || !ok {
		return fmt.Errorf("Invalid Key (%s) for Message structure", key)
	}

	kind, _ := reflections.GetFieldKind(m, key)
	err = fmt.Errorf("Unexpected type %s for SetField", kind)
	switch kind {
	case reflect.String:
		err = reflections.SetField(m, key, value)
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(value)
		err = reflections.SetField(m, key, b)
	case reflect.Int:
		var i int
		i, err = strconv.Atoi(value)
		err = reflections.SetField(m, key, i)
	}
	return err
}

func (m *Message) Output() (output string, err error) {
	if m.Capability == 0 {
		return "", errors.New("Capability not defined")
	}

	ok := false
	if m.capabilityMessage, ok = messageCodes[m.Capability]; !ok {
		return "", fmt.Errorf("Invalid message capability (%d)", m.Capability)
	}
	output += fmt.Sprintf("%d %s\n", m.Capability, m.capabilityMessage)

	structTags, _ := reflections.Tags(m, "output")
	for k, o := range structTags {
		if o == "" {
			continue
		}
		kind, _ := reflections.GetFieldKind(m, k)
		switch kind {
		case reflect.String:
			v, err := reflections.GetField(m, k)
			if err != nil {
				return "", err
			}

			if v.(string) != "" {
				output += fmt.Sprintf("%s: %s\n", o, v.(string))
			}
		case reflect.Bool:
			v, err := reflections.GetField(m, k)
			if err != nil {
				return "", err
			}

			if v.(bool) {
				output += fmt.Sprintf("%s: true\n", o)
			}
		case reflect.Int:
			v, err := reflections.GetField(m, k)
			if err != nil {
				return "", err
			}

			if v.(int) != 0 {
				output += fmt.Sprintf("%s: %d\n", o, v)
			}
		case reflect.Ptr:
			ptrValue, _ := reflections.GetField(m, k)
			switch v := ptrValue.(type) {
			case *time.Time:
				if v != nil {
					output += fmt.Sprintf("%s: %s\n", o, v.Format(time.RFC1123))
				}
			default:
				fmt.Fprintf(os.Stderr, "Unrecognised Messsage Field Ptr Kind for %s: %v\n", k, kind)
			}

		default:
			fmt.Fprintf(os.Stderr, "Unrecognised Messsage Field Kind for %s: %v\n", k, kind)
		}
	}

	return output, err
}
