package driver

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/thomasmitchell/shuttle-resource/models"
)

type s3Driver struct {
	s3     *s3.S3
	bucket string
	path   string
}

func newS3Driver(cfg models.Source) (*s3Driver, error) {
	session := session.Must(session.NewSession())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TLSSkipVerify,
			},
		},
	}

	awsCfg := aws.NewConfig().
		WithCredentials(
			credentials.NewStaticCredentials(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		).
		WithEndpoint(cfg.Endpoint).
		WithRegion(cfg.Region).
		WithHTTPClient(client)

	return &s3Driver{
			s3:     s3.New(session, awsCfg),
			bucket: cfg.Bucket,
			path:   strings.TrimSuffix(cfg.Path, "/") + "/",
		},
		nil
}

func (s *s3Driver) Read(version models.Version) (*models.Payload, error) {
	key := s.keyFor(version)
	getObjOut, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		ioutil.ReadAll(getObjOut.Body)
		getObjOut.Body.Close()
	}()

	dec := json.NewDecoder(getObjOut.Body)
	ret := models.Payload{}
	err = dec.Decode(&ret)
	return &ret, err
}

func (s *s3Driver) Write(version models.Version, payload models.Payload) error {
	jBuf, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	key := s.keyFor(version)

	_, err = s.s3.PutObject(&s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   bytes.NewReader(jBuf),
	})

	return err
}

func (s *s3Driver) LatestVersion() (models.Version, error) {
	ret := models.NewVersion("0")
	//there... really should never be more than 1000 active versions in the bucket
	// unless something has gone wrong, so I've forgone pagination for now.
	listObjOut, err := s.s3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &s.path,
	})
	if err != nil {
		return ret, err
	}

	for _, obj := range listObjOut.Contents {
		candidate := models.NewVersion(filepath.Base(*obj.Key))
		if ret.LessThan(candidate) {
			ret = candidate
		}
	}

	return ret, nil
}

func (s *s3Driver) Clean(version models.Version) error {
	key := s.keyFor(version)
	_, err := s.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	return err
}

func (s *s3Driver) keyFor(version models.Version) string {
	return s.path + version.Number
}
