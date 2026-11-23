package main

import (
	"encoding/json"
	"fmt"
	"github.com/ppc64le-cloud/build-bot/pkg/storage/s3"
	"os"
)

const (
	myBucket = "ppc64le-ci-builds"
)

var (
	sess *s3.S3
)

func backingStoreInit() error {
	if s3CredentialsFile == "" {
		return fmt.Errorf("--s3-credentials-file is missing")
	}
	s3Credentials, err := os.ReadFile(s3CredentialsFile)
	if err != nil {
		return fmt.Errorf("failed to read s3 credentials file %q: %w", s3CredentialsFile, err)
	}

	var cred s3.Credentials
	err = json.Unmarshal(s3Credentials, &cred)
	if err != nil {
		return fmt.Errorf("failed to parse s3 credentials file %q: %w", s3CredentialsFile, err)
	}
	sess, err = s3.NewSession(&cred)
	if err != nil {
		return fmt.Errorf("failed to create storage session: %w", err)
	}
	return nil
}
