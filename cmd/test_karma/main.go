package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const urlTemplate = "http://127.0.0.1:8080/api/v1/file?filename=%s"

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	genBuf := make([]byte, 10*1024*1024)
	if _, err := r.Read(genBuf); err != nil {
		log.Fatalf("error while generating random string: %s", err)
	}

	firstTryFileName := "testfile"
	if err := postFile(http.DefaultClient, firstTryFileName, bytes.NewReader(genBuf)); err != nil {
		log.Fatalf("error posting file: %s", err)
	}

	downloadedFirstFile, err := getFile(http.DefaultClient, firstTryFileName)
	if err != nil {
		log.Fatalf("error downloading file: %s", err)
	}

	if bytes.Equal(genBuf, downloadedFirstFile) {
		log.Println("Generated and downloaded firstfile are equal")
	}

	// wait for additional storage node
	<-time.After(1 * time.Minute)
	secondTryFileName := "testfile2"
	if err := postFile(http.DefaultClient, secondTryFileName, bytes.NewReader(genBuf)); err != nil {
		log.Fatalf("error posting file: %s", err)
	}

	downloadedSecondFile, err := getFile(http.DefaultClient, secondTryFileName)
	if err != nil {
		log.Fatalf("error downloading file: %s", err)
	}

	if bytes.Equal(downloadedSecondFile, downloadedFirstFile) {
		log.Println("Downloaded first and second files are equal")
	}
}

func postFile(client *http.Client, filename string, data io.Reader) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf(urlTemplate, filename), data)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func getFile(client *http.Client, filename string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(urlTemplate, filename), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request: %s", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(fmt.Errorf("body close error: %w", err).Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	return buf, nil
}
