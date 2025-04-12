package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func Run(command string, in string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := make(chan string)
	defer close(c)
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if len(in) > 0 {
		cmd.Stdin = strings.NewReader(in)
	}
	err := cmd.Run()
	if err != nil {
		return "", cmd.String(), err
	}
	if stderr.String() != "" {
		log.Println(stderr.String())
		return "", cmd.String(), fmt.Errorf(stderr.String())
	}

	res := strings.Replace(stdout.String(), "\n", "", -1)
	return res, cmd.String(), nil
}
