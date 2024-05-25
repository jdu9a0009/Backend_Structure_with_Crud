package hashing

import (
	"fmt"
	"time"
)

func GenerateSlug(s string) string {
	t := time.Now().Unix()

	var slug string
	a := 0
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			a = 0
			slug += string(v + 32)
		} else if v >= 'a' && v <= 'z' {
			a = 0
			slug += string(v)
		} else if v >= '0' && v <= '9' {
			a = 0
			slug += string(v)
		} else if v == ',' {
			a = 1
			slug += "-"
		} else if v == ' ' && a == 0 {
			slug += "-"
		} else if v == '-' {
			a = 0
			slug += "-"
		}
	}

	slug += fmt.Sprintf("-%d", t)

	return slug
}
