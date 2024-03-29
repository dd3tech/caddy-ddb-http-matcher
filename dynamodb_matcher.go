package dynamodb_matcher

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	UrlIndex    int    `json:"url_index"`
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

// Match returns true if the key-value pair exists in the specified DynamoDB table.
func (m DynamoDBMatcher) Match(r *http.Request) bool {

	valueCheck := strings.Split(r.Host, ".")[m.UrlIndex]

	input := &dynamodb.GetItemInput{
		TableName: &m.TableName,
		Key: map[string]types.AttributeValue{
			m.KeyName: &types.AttributeValueMemberS{Value: valueCheck},
		},
	}

	result, err := m.dynamoDBSvc.GetItem(context.TODO(), input)
	if err != nil {
		// Log error or handle accordingly
		fmt.Println(err)
		return false
	}

	if result.Item != nil {
		fmt.Println("Website Migrated: ", valueCheck)
		return true
	}

	// Item does not exist, it's not a match
	return false

}

// CaddyModule returns the Caddy module information.
func (DynamoDBMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.ddb_matcher",
		New: func() caddy.Module { return new(DynamoDBMatcher) },
	}
}

// UnmarshalCaddyfile sets up the matcher from Caddyfile tokens. Syntax: dynamodb <table_name> <key_name> <key_value>
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
		case "url_index":
			d.NextArg()
			UrlIndex, err := strconv.Atoi(d.Val())
			if err != nil {
				return nil
			}
			dm.UrlIndex = UrlIndex
		case "region":
			d.NextArg()
			dm.Region = d.Val()
		}
	}

	return nil
}

// Interface guards
var (
	_ caddy.Validator          = (*DynamoDBMatcher)(nil)
	_ caddyhttp.RequestMatcher = (*DynamoDBMatcher)(nil)
)

// Validate ensures that the matcher is properly configured.
func (dm *DynamoDBMatcher) Validate() error {
	// if dm.TableName == "" || dm.KeyName == "" {
	// 	return errors.New("DynamoDB parameters cannot be empty")
	// }
	return nil
}

func init() {
	caddy.RegisterModule(DynamoDBMatcher{})
}
