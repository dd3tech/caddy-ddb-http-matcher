# Caddy DynamoDB HTTP Matcher

This module for the Caddy web server allows you to match HTTP requests based on the results of queries to AWS DynamoDB. It can be used to conditionally apply Caddy directives based on data stored in DynamoDB.

It's an easiest way to route traffic depending on a dynamodb query match

## Features

- **DynamoDB Integration**: Seamlessly query DynamoDB tables to make routing decisions.
- **Flexible Matching**: Use any DynamoDB table and specify the key-value pairs for matching.
- **Easy Configuration**: Configure directly from your Caddyfile.

## Prerequisites

- Caddy 2.x
- AWS Account with access to DynamoDB

## Installation

To use this module, you need to build Caddy with this custom module included. Follow these steps:

1. **Get the module**:

    ```bash
    go get github.com/dd3tech/caddy-ddb-http-matcher.git
    ```

2. **Build Caddy with the module**:

    Check out the [Caddy developer guide](https://caddyserver.com/docs/extending-caddy) on how to include external modules in your build.

    You can do it with xcaddy

    ```
    xcaddy build --with github.com/dd3tech/caddy-ddb-http-matcher.git
    ./caddy run
    ```


## Configuration

Add the matcher configuration to your Caddyfile. Here's an example:

```Caddyfile
{
    order dynamodb before path
}

:80 {
    @isActive ddb_matcher {
        table_name "your_table_name"
        key_name "your_key"
        Url_index "idx_to_match_in_host"
        access_key "aws_access_key"
        secret_key "aws_secret_key"
        region "aws_region"
    }
    handle @isActive {
        reverse_proxy http://siteA.com
    }
    handle reverse_proxy http://siteB.com
}
