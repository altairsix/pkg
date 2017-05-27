package local

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/altairsix/pkg/context"
	"github.com/altairsix/pkg/types"
	"github.com/altairsix/pkg/web/session"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	Region = "us-west-2"
	Env    = "local"
)

var (
	Context  context.Kontext = context.Background(Env)
	DynamoDB *dynamodb.DynamoDB
	SNS      *sns.SNS
	SQS      *sqs.SQS
)

var (
	SessionStore sessions.Store
)

var (
	IDFactory types.IDFactory
)

func init() {
	// Read Env
	//
	dir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln(err)
	}

	env := map[string]string{}
	for i := 0; i < 8; i++ {
		dir = filepath.Join(dir, "..")
		filename := filepath.Join(dir, "test.json")
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}

		err = json.Unmarshal(data, &env)
		if err != nil {
			log.Fatalln(err)
		}
		break
	}

	for k, v := range env {
		os.Setenv(k, v)
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-west-2"
	}

	// Configure AWS
	//
	cfg := &aws.Config{Region: aws.String(region)}
	s, err := aws_session.NewSession(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	DynamoDB = dynamodb.New(s)
	SNS = sns.New(s)
	SQS = sqs.New(s)

	// Setup IDFactory
	//
	id := time.Now().UnixNano()
	IDFactory = func() types.ID {
		atomic.AddInt64(&id, 1)
		return types.ID(id)
	}

	// SessionStore
	//
	codecs, err := session.EnvCodecs()
	if err != nil {
		hashKey := securecookie.GenerateRandomKey(64)
		blockKey := securecookie.GenerateRandomKey(32)

		codecs = []securecookie.Codec{securecookie.New(hashKey, blockKey)}
	}

	SessionStore = &sessions.CookieStore{
		Codecs: codecs,
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
	}
}

func NewID() types.ID {
	return IDFactory.NewID()
}

func NewKey() types.Key {
	return IDFactory.NewKey()
}
