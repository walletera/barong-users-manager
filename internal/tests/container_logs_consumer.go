package tests

import (
	"bufio"
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go"
)

type ContainerLogConsumer struct {
	file *os.File
}

func NewContainerLogConsumer(containerName string) *ContainerLogConsumer {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	logDir := fmt.Sprintf("%s/containerlogs", wd)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}
	file, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, containerName))
	if err != nil {
		panic(err)
	}
	return &ContainerLogConsumer{file: file}
}

func (c *ContainerLogConsumer) Accept(log testcontainers.Log) {
	w := bufio.NewWriter(c.file)
	if _, err := w.Write(log.Content); err != nil {
		panic(err)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
}
