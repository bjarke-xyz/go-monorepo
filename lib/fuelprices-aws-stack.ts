import { Duration, RemovalPolicy, Stack, StackProps } from "aws-cdk-lib";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import { Certificate } from "aws-cdk-lib/aws-certificatemanager";
import { AttributeType, Table } from "aws-cdk-lib/aws-dynamodb";
import { Rule, Schedule } from "aws-cdk-lib/aws-events";
import { LambdaFunction } from "aws-cdk-lib/aws-events-targets";
import { Runtime } from "aws-cdk-lib/aws-lambda";
import {
  S3EventSource,
  SqsEventSource,
} from "aws-cdk-lib/aws-lambda-event-sources";
import { NodejsFunction } from "aws-cdk-lib/aws-lambda-nodejs";
import * as s3 from "aws-cdk-lib/aws-s3";
import { EventType } from "aws-cdk-lib/aws-s3";
import { Queue } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import * as path from "path";

export class FuelpricesAwsStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const bucket = new s3.Bucket(this, "FuelpriceStore", {
      removalPolicy: RemovalPolicy.DESTROY,
    });

    const dynamoTable = new Table(this, "prices", {
      partitionKey: {
        name: "PK",
        type: AttributeType.STRING,
      },
      sortKey: {
        name: "SK",
        type: AttributeType.STRING,
      },
      tableName: "prices",
      removalPolicy: RemovalPolicy.DESTROY,
      readCapacity: 4,
      writeCapacity: 4,
    });

    const queue = new Queue(this, "price-chunk-buffer");

    const getFuelpriceHandler = new NodejsFunction(
      this,
      "get-fuelprice-handler",
      {
        memorySize: 128,
        timeout: Duration.seconds(5),
        runtime: Runtime.NODEJS_14_X,
        handler: "main",
        entry: path.join(__dirname, "../src/lambdas/get-fuelprice-handler.ts"),
        environment: {
          BUCKET: bucket.bucketName,
          TABLE_NAME: dynamoTable.tableName,
          NODE_ENV: "production",
        },
        bundling: {
          minify: true,
          externalModules: ["aws-sdk"],
        },
      }
    );
    bucket.grantRead(getFuelpriceHandler);
    dynamoTable.grantReadData(getFuelpriceHandler);

    const cacheRefreshHandler = new NodejsFunction(
      this,
      "cache-refresh-handler",
      {
        memorySize: 128,
        timeout: Duration.seconds(60),
        runtime: Runtime.NODEJS_14_X,
        handler: "main",
        entry: path.join(__dirname, "../src/lambdas/cache-refresh-handler.ts"),
        environment: {
          BUCKET: bucket.bucketName,
          TABLE_NAME: dynamoTable.tableName,
          SQS_URL: queue.queueUrl,
          NODE_ENV: "production",
        },
        bundling: {
          minify: true,
          externalModules: ["aws-sdk"],
        },
      }
    );
    bucket.grantRead(cacheRefreshHandler);
    dynamoTable.grantReadWriteData(cacheRefreshHandler);
    queue.grantSendMessages(cacheRefreshHandler);

    const dataFetcherHandler = new NodejsFunction(
      this,
      "data-fetcher-handler",
      {
        memorySize: 128,
        timeout: Duration.seconds(15),
        runtime: Runtime.NODEJS_14_X,
        handler: "main",
        entry: path.join(__dirname, "../src/lambdas/data-fetcher-handler.ts"),
        environment: {
          BUCKET: bucket.bucketName,
          NODE_ENV: "production",
        },
        bundling: {
          minify: true,
          externalModules: ["aws-sdk"],
        },
      }
    );
    bucket.grantReadWrite(dataFetcherHandler);

    const cacheWriteHandler = new NodejsFunction(this, "cache-write-handler", {
      memorySize: 128,
      timeout: Duration.seconds(30),
      runtime: Runtime.NODEJS_14_X,
      handler: "main",
      entry: path.join(__dirname, "../src/lambdas/cache-write-handler.ts"),
      environment: {
        TABLE_NAME: dynamoTable.tableName,
        SQS_URL: queue.queueUrl,
        NODE_ENV: "production",
      },
      bundling: {
        minify: true,
        externalModules: ["aws-sdk"],
      },
    });
    queue.grantConsumeMessages(cacheWriteHandler);
    dynamoTable.grantReadWriteData(cacheWriteHandler);

    const apiGateway = new apigateway.RestApi(this, "fuelprices-api", {
      restApiName: "Fuelprices API",
      description: "API for getting fuel prices",
      domainName: {
        certificate: Certificate.fromCertificateArn(
          this,
          "df0ce0af-7bc4-4d78-b583-3e43b947e842",
          "arn:aws:acm:eu-north-1:573355056124:certificate/df0ce0af-7bc4-4d78-b583-3e43b947e842"
        ),
        domainName: "fuelprices.bjarke.xyz",
      },
    });

    const getFuelpriceIntegration = new apigateway.LambdaIntegration(
      getFuelpriceHandler,
      {
        requestTemplates: { "application/json": '{"statusCode": "200"}' },
      }
    );

    const cronRule = new Rule(this, "CronRule", {
      schedule: Schedule.expression("rate(1 hour)"),
    });
    cronRule.addTarget(new LambdaFunction(dataFetcherHandler));

    cacheRefreshHandler.addEventSource(
      new S3EventSource(bucket, {
        events: [EventType.OBJECT_CREATED_PUT],
      })
    );

    cacheWriteHandler.addEventSource(new SqsEventSource(queue));

    apiGateway.root.addMethod("GET", getFuelpriceIntegration, {});
  }
}
