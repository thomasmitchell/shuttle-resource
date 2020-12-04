package driver

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
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

func (s *s3Driver) Versions() (models.VersionList, error) {
	ret := models.VersionList{}
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
		vString := filepath.Base(*obj.Key)
		v, err := models.VersionFromString(vString)
		if err != nil {
			return nil,
				fmt.Errorf(
					"Version could not be parsed from string `%s': %s",
					vString,
					err,
				)
		}

		ret = append(ret, v)
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
	return s.path + version.String()
}
