# AWS

## DynamoDB

See `Populate` section below.

## Populate

The easiest way to populate the findingaid is to use the `create-dynamodb-import` tool to create a CSV file of all the findingaid pointers which can be used to create (and populate) a new "findingaid" table in DynamoDB.

```
$> ./bin/create-dynamodb-import /usr/local/data/whosonfirst-findingaids/data/* /usr/local/data/whosonfirst-findingaids-venue/data/* > findingaid.csv
$> gzip findingaid.csv
$> aws --profile {PROFILE} s3 cp findingaid.csv.gz s3://{BUCKET}
```

Follow the instructions for importing the CSV file, specifying a new table called "findingaid" with a partition key of "id" which is a number ("N").

* https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataImport.Requesting.html#S3DataImport.Requesting.Console

Importing 23,310,204 findingaid pointers in September, 2022 took about 75 minutes.

_To do: Add notes about populating a standalone instance of DynamoDB outside of AWS._

## Sync / update

```
$> cd /usr/local/whosonfirst/whosonfirst-findingaids
$> make docker
```

### ECS

_ECS documentation is incomplete._

#### Tasks

```
/usr/local/bin/update-findingaids.sh,-T,awsparamstore://whosonfirst-findingaid-github-token?region=us-east-1&credentials=iam:,-O,3600
```

### IAM

#### Policies

##### FindingaidDynamoECS

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ListAndDescribe",
            "Effect": "Allow",
            "Action": [
                "dynamodb:List*",
                "dynamodb:DescribeReservedCapacity*",
                "dynamodb:DescribeLimits",
                "dynamodb:DescribeTimeToLive"
            ],
            "Resource": "*"
        },
        {
            "Sid": "SpecificTable",
            "Effect": "Allow",
            "Action": [
                "dynamodb:BatchGet*",
                "dynamodb:DescribeStream",
                "dynamodb:DescribeTable",
                "dynamodb:Get*",
                "dynamodb:Query",
                "dynamodb:Scan",
                "dynamodb:BatchWrite*",
                "dynamodb:Update",
                "dynamodb:PutItem"
            ],
            "Resource": "arn:aws:dynamodb:*:*:table/findingaid"
        }
    ]
}
```

##### WOFParameterStoreFindingAidGithub

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "ssm:DescribeParameters"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": "ssm:GetParameter",
            "Resource": "arn:aws:ssm:{REGION}:{ACCOUNT}:parameter/whosonfirst-findingaid-github-token"
        },
        {
            "Effect": "Allow",
            "Action": [
                "kms:Decrypt"
            ],
            "Resource": [
                "arn:aws:kms:{REGION}:{ACCOUNT}:key/CMK"
            ]
        }
    ]
}
```

#### Roles

##### WOFECSFindingAid

* FindingaidDynamoECS
* WOFParameterStoreFindingAidGithub
* AmazonECSTaskExecutionRolePolicy

### EventBridge

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
