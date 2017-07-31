package s3apt

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"hash"
	"io"
	"net/url"
	"os"
)

type Session struct {
	session *session.Session
	service *s3.S3
	context context.Context
}

func NewS3Apt() *Session {

	regionSession, _ := session.NewSession()
	region := *regionSession.Config.Region
	if region == "" {
		metaClient := ec2metadata.New(regionSession)
		region, _ = metaClient.Region()
	}

	s := &Session{
		session: session.Must(session.NewSession(aws.NewConfig().WithRegion(region))),
		context: context.Background(),
	}

	// Create the service's client with the session.
	s.service = s3.New(s.session)

	return s
}

func parseBucketKey(uri string) (bucket string, key string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	bucket = u.Host
	key = u.Path
	return
}

func computeHashes(filePath string) (hashMD5, hashSHA256, hashSHA512 string, err error) {

	var results []string
	for _, h := range []hash.Hash{md5.New(), sha256.New(), sha512.New()} {
		var r []byte
		var f *os.File
		f, err = os.Open(filePath)
		if err != nil {
			return
		}
		defer f.Close()
		if _, err = io.Copy(h, f); err != nil {
			return
		}
		results = append(results, fmt.Sprintf("%x", h.Sum(r)))
	}

	return results[0], results[1], results[2], nil
}

func (s *Session) SetRegion(uri string) (err error) {
	bucket, _, err := parseBucketKey(uri)
	if err != nil {
		return
	}

	regionHint := *s.session.Config.Region
	if envHint, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		regionHint = envHint
	}
	if envHint, ok := os.LookupEnv("AWS_REGION"); ok {
		regionHint = envHint
	}

	region, err := s3manager.GetBucketRegion(s.context, s.session, bucket, regionHint)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return fmt.Errorf("unable to find bucket %s's region not found\n", bucket)
		}
		if aerr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("AWS Error: %s \n", aerr.Message())
		}
		return err
	}
	s.session.Config.Region = aws.String(region)
	s.service.Config.Region = aws.String(region)

	return nil
}

func (s *Session) Get(uri string, filename string, method *Method) (err error) {
	bucket, key, err := parseBucketKey(uri)
	if err != nil {
		return err
	}

	method.Send(&Message{
		Capability: 102,
		Uri:        uri,
		Message:    "Waiting for Headers",
	})

	err = s.SetRegion(uri)
	if err != nil {
		return err
	}

	results, err := s.service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get file, %v", err)
	}
	defer results.Body.Close()

	method.Send(&Message{
		Capability:   200,
		LastModified: results.LastModified,
		Uri:          uri,
		Size:         int(*results.ContentLength),
	})

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %q, %v", filename, err)
	}
	defer f.Close()

	n, err := io.Copy(f, results.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file %q, %v", filename, err)
	}

	md5sum, sha256sum, sha512sum, err := computeHashes(filename)
	if err != nil {
		return err
	}

	method.Send(&Message{
		Capability:   201,
		LastModified: results.LastModified,
		Uri:          uri,
		Size:         int(n),
		Filename:     filename,
		MD5Hash:      md5sum,
		MD5SumHash:   md5sum,
		SHA256Hash:   sha256sum,
		SHA512Hash:   sha512sum,
	})

	return nil

}
