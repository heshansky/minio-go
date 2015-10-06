package minio

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const expirationFormat = "2006-01-02T15:04:05.999Z"

// contentLengthRange - min and max size of content
type contentLengthRange struct {
	min int
	max int
}

func (c contentLengthRange) marshalJSON() string {
	return fmt.Sprintf("[\"content-length-range\",%d,%d]", c.min, c.max)
}

// Policy explanation: http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-HTTPPOSTConstructPolicy.html
type policy struct {
	matchType string
	key       string
	value     string
}

func (p policy) marshalJSON() string {
	return fmt.Sprintf("[\"%s\",\"%s\",\"%s\"]", p.matchType, p.key, p.value)
}

// PostPolicy provides strict static type conversion and validation for Amazon S3's POST policy JSON string.
type PostPolicy struct {
	expiration         time.Time // expiration date and time of the POST policy.
	policies           []policy
	contentLengthRange contentLengthRange

	// Post form data
	formData map[string]string
}

func NewPostPolicy() *PostPolicy {
	p := &PostPolicy{}
	p.policies = make([]policy, 0)
	p.formData = make(map[string]string)
	return p
}

// SetExpires expiration time
func (p *PostPolicy) SetExpires(t time.Time) error {
	if t.IsZero() {
		return errors.New("time input invalid")
	}
	p.expiration = t
	return nil
}

// SetKey Object name
func (p *PostPolicy) SetKey(key string) error {
	if len(key) == 0 {
		return errors.New("key invalid")
	}
	policy := policy{"eq", "$key", key}
	p.policies = append(p.policies, policy)
	p.formData["key"] = key
	return nil
}

// SetKeyStartsWith Object name that can start with
func (p *PostPolicy) SetKeyStartsWith(keyStartsWith string) error {
	if len(keyStartsWith) == 0 {
		return errors.New("key-starts-with invalid")
	}
	policy := policy{"starts-with", "$key", keyStartsWith}
	p.policies = append(p.policies, policy)
	p.formData["key"] = keyStartsWith
	return nil
}

// SetBucket bucket name
func (p *PostPolicy) SetBucket(bucket string) error {
	if len(bucket) == 0 {
		return errors.New("bucket invalid")
	}
	policy := policy{"eq", "$bucket", bucket}
	p.policies = append(p.policies, policy)
	p.formData["bucket"] = bucket
	return nil
}

// SetContentType content-type
func (p *PostPolicy) SetContentType(contentType string) error {
	if len(contentType) == 0 {
		return errors.New("contentType invalid")
	}
	policy := policy{"eq", "$Content-Type", contentType}
	p.policies = append(p.policies, policy)
	p.formData["Content-Type"] = contentType
	return nil
}

// MarshalJSON provides Marshalled JSON
func (p PostPolicy) marshalJSON() []byte {
	expirationstr := ""
	if p.expiration.IsZero() == false {
		expirationstr = `"expiration":"` + p.expiration.Format(expirationFormat) + `"`
	}
	policiesstr := ""
	policies := []string{}
	for _, policy := range p.policies {
		policies = append(policies, policy.marshalJSON())
	}
	if p.contentLengthRange.min != 0 || p.contentLengthRange.max != 0 {
		policies = append(policies, p.contentLengthRange.marshalJSON())
	}
	if len(policies) > 0 {
		policiesstr = `"conditions":[` + strings.Join(policies, ",") + "]"
	}
	retstr := "{"
	if len(expirationstr) > 0 {
		retstr = retstr + expirationstr + ","
	}
	if len(policiesstr) > 0 {
		retstr = retstr + policiesstr
	}
	retstr = retstr + "}"
	return []byte(retstr)
}

// Base64 Base64() of PostPolicy's Marshalled json
func (p PostPolicy) base64() string {
	b := p.marshalJSON()
	return base64.StdEncoding.EncodeToString(b)
}
