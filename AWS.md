# AWS

## DynamoDB

Table called "findingaid" with a partition key of "id" which is a number ("N").

## Populate

## Sync

## Resolver(d)

### Lambda

Create a new Lambda function called `FindingaidResolverServer` and upload the following code:

```
$> make lambda
if test -f main; then rm -f main; fi
if test -f resolverd.zip; then rm -f resolverd.zip; fi
GOOS=linux go build -mod vendor -o main cmd/resolverd/main.go
zip resolverd.zip main
  adding: main (deflated 55%)
rm -f main
```

#### Environment variables

| Key | Value | Notes | 
| RESOLVERD_RESOLVER_URI | awsdynamodb://findingaid?partition_key=id&region={REGION}&credentials=iam: | |
| RESOLVERD_SERVER_URI | lambda:// | |

### IAM

#### Policies

##### FindingaidResolverServerDynamoDB

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "FindingAid",
            "Effect": "Allow",
            "Action": [
                "dynamodb:BatchGet*",
                "dynamodb:DescribeStream",
                "dynamodb:DescribeTable",
                "dynamodb:Get*",
                "dynamodb:Query",
                "dynamodb:Scan"
            ],
            "Resource": "arn:aws:dynamodb:*:*:table/findingaid"
        }
    ]
}
```

#### Roles

##### WhosOnFirstLambdaFindingaidResolverServer

* AWSLambdaBasicExecutionRole
* FindingaidResolverServerDynamoDB

### API Gateway

Create a new "REST" API and configure it with a new `{proxy+}` resources.

Delete the `ANY` method and then create a new `GET` method and configure it to point to the `FindingaidResolverServer` Lambda function.

Deploy the new API with a new stage name called "findingaid". When testing you should see something like this (assuming the findingaid hasn't been populated yet:

```
$> curl -s https://{PREFIX}.execute-api.{REGION}.amazonaws.com/findingaid/id/123456
Failed to get record for 123456, item {map[id:123456 repo_name:] map[id:123456 repo_name:] {<nil> <nil> 0} []} not found (code=NotFound)
```