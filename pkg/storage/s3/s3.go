package s3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ppc64le-cloud/build-bot/pkg/storage"
	"k8s.io/klog/v2"
)

var _ storage.Storage = &S3{}

const (
	defaultPartSize    = 16 * 1024 * 1024 // 16 MB
	defaultConcurrency = 5
)

type S3 struct {
	sess       *session.Session
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

type Credentials struct {
	Region           string `json:"region"`
	Endpoint         string `json:"endpoint"`
	Insecure         bool   `json:"insecure"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
	AccessKey        string `json:"access_key"`
	SecretKey        string `json:"secret_key"`
}

// Validate performs structural checks on the configuration. It does not
// verify that the credentials are accepted by the S3 endpoint.
func (c *Credentials) Validate() error {
	if c.AccessKey == "" || c.SecretKey == "" {
		return errors.New("s3: access_key and secret_key are both required in the credentials file")
	}
	if c.Region == "" {
		return errors.New("s3: region is required")
	}
	if c.Endpoint == "" {
		return errors.New("s3: endpoint is required")
	}
	if u, err := url.Parse(c.Endpoint); err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("s3: invalid endpoint %q, expected a URL like https://s3.us-south.cloud-object-storage.appdomain.cloud", c.Endpoint)
	}
	return nil
}

func newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 32
	transport.IdleConnTimeout = 90 * time.Second

	return &http.Client{
		Transport: transport,
	}
}

func NewSession(cfg *Credentials) (*S3, error) {
	if cfg == nil {
		return nil, errors.New("s3: configuration credentials cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Validate() guarantees region, endpoint, access_key and secret_key are set.
	awsCfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		DisableSSL:       aws.Bool(cfg.Insecure),
		Endpoint:         aws.String(cfg.Endpoint),
		HTTPClient:       newHTTPClient(),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(cfg.S3ForcePathStyle),
	}

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, fmt.Errorf("s3: failed to create session: %w", err)
	}

	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = defaultPartSize
		u.Concurrency = defaultConcurrency
	})

	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = defaultPartSize
		d.Concurrency = defaultConcurrency
	})

	return &S3{
		sess:       sess,
		client:     s3.New(sess),
		uploader:   uploader,
		downloader: downloader,
	}, nil
}

func (s *S3) Download(bucket, object string) ([]byte, error) {
	wb := aws.WriteAtBuffer{}
	n, err := s.downloader.Download(&wb, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download %s/%s: %w", bucket, object, err)
	}
	klog.Infof("file downloaded, %d bytes", n)
	return wb.Bytes(), nil
}

func (s *S3) Upload(bucket, object string, content []byte) error {
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(http.DetectContentType(content)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s/%s: %w", bucket, object, err)
	}
	klog.Infof("file uploaded to, %s", result.Location)
	return nil
}

func (s *S3) Head(bucket, object string) (*storage.ObjectInfo, error) {
	out, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to head %s/%s: %w", bucket, object, err)
	}

	info := &storage.ObjectInfo{}
	if out.ContentLength != nil {
		info.ContentLength = *out.ContentLength
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	return info, nil
}

func (s *S3) Get(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
	out, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object %s/%s: %w", bucket, object, err)
	}

	info := &storage.ObjectInfo{}
	if out.ContentLength != nil {
		info.ContentLength = *out.ContentLength
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}

	return out.Body, info, nil
}
