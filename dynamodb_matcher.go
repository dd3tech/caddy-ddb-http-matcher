package dynamodb_matcher

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

// DynamoDBMatcher is a Caddy matcher that checks for the existence of a key-value pair in a DynamoDB table.
type DynamoDBMatcher struct {
	TableName   string `json:"table_name"`
	KeyName     string `json:"key_name"`
	Regex       string `json:"url_index"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Region      string `json:"region"`
	dynamoDBSvc *dynamodb.Client
}

// Provision sets up the module's configuration.
func (m *DynamoDBMatcher) Provision(ctx caddy.Context) error {
	creds := credentials.NewStaticCredentialsProvider(m.AccessKey, m.SecretKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(m.Region))
	if err != nil {
		return err
	}

	m.dynamoDBSvc = dynamodb.NewFromConfig(cfg)
	return nil
}

func (m DynamoDBMatcher) Match(r *http.Request) bool {

	var valueCheck string

	rgx, err := regexp.Compile(m.Regex)
	patternWww := regexp.MustCompile(`(?:www)`)

	if err != nil {
		fmt.Println("Error compiling regex: ", err.Error())
		return false
	}

	if patternWww.MatchString(r.Host) {
		// get rid off www
		valueCheck = strings.Split(rgx.FindString(r.Host), ".")[1]
	} else {
		valueCheck = rgx.FindString(r.Host)
	}

	log.Println("Value to check - DynamoDB Matcher: ", valueCheck)

	input := &dynamodb.GetItemInput{
		TableName: &m.TableName,
		Key: map[string]types.AttributeValue{
			m.KeyName: &types.AttributeValueMemberS{Value: valueCheck},
		},
	}

	result, err := m.dynamoDBSvc.GetItem(context.TODO(), input)
	if err != nil {
		// Log error or handle accordingly
		fmt.Println("Got error calling GetItem: ", err.Error())
		return false
	}

	if result.Item != nil {
		return true
	}

	return false

}

func (DynamoDBMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.ddb_matcher",
		New: func() caddy.Module { return new(DynamoDBMatcher) },
	}
}

func (dm *DynamoDBMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()

	for d.Next() {
		switch d.Val() {
		case "table_name":
			d.NextArg()
			dm.TableName = d.Val()
		case "key_name":
			d.NextArg()
			dm.KeyName = d.Val()
		case "access_key":
			d.NextArg()
			dm.AccessKey = d.Val()
		case "secret_key":
			d.NextArg()
			dm.SecretKey = d.Val()
		case "regex":
			d.NextArg()
			dm.Regex = d.Val()
		case "region":
			d.NextArg()
			dm.Region = d.Val()
		}
	}

	return nil
}

var (
	_ caddy.Validator          = (*DynamoDBMatcher)(nil)
	_ caddyhttp.RequestMatcher = (*DynamoDBMatcher)(nil)
)

func (dm *DynamoDBMatcher) Validate() error {
	if dm.TableName == "" || dm.KeyName == "" || dm.AccessKey == "" || dm.SecretKey == "" || dm.Region == "" {
		return errors.New("DynamoDB parameters missed. Please check the configuration file")
	}
	return nil
}

func init() {
	caddy.RegisterModule(DynamoDBMatcher{})
}
