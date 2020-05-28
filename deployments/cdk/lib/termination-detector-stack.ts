import * as cdk from '@aws-cdk/core';
import * as apiGW from '@aws-cdk/aws-apigateway';
import * as lambda from '@aws-cdk/aws-lambda';
import * as path from "path";
import * as dynamo from '@aws-cdk/aws-dynamodb';

export class TerminationDetectorStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const tasksTable = new dynamo.Table(this, 'tasks-table', {
      partitionKey: {name: 'process_id', type: dynamo.AttributeType.STRING},
      sortKey: {name: 'task_id', type: dynamo.AttributeType.STRING},
      billingMode: dynamo.BillingMode.PAY_PER_REQUEST,
      timeToLiveAttribute: 'ttl',
    });
    tasksTable.addLocalSecondaryIndex({
      indexName: 'badStateEnterTimeIndex',
      sortKey: {name: 'bad_state_enter_time', type: dynamo.AttributeType.STRING},
      projectionType: dynamo.ProjectionType.ALL,
    })

    const apiLambda = new lambda.Function(this, 'api-lambda', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'api',
      code: lambda.Code.fromAsset(path.join(__dirname, '..', '..', '..', 'build', 'api.zip')),
      environment: {
        TASKS_TABLE_NAME: tasksTable.tableName,
        TASKS_STORING_DURATION: '168h'
      }
    });
    tasksTable.grantReadWriteData(apiLambda);

    const apiLambdaIntegration = new apiGW.LambdaIntegration(apiLambda)

    const api = new apiGW.RestApi(this, 'processes-api');

    const processes = api.root.addResource('processes');
    const process = processes.addResource('{process_id}');
    process.addMethod('GET', apiLambdaIntegration, {
      authorizationType: apiGW.AuthorizationType.IAM,
    })
    const tasks = process.addResource('tasks');
    const task = tasks.addResource('{task_id}');
    task.addMethod('PUT', apiLambdaIntegration, {
      authorizationType: apiGW.AuthorizationType.IAM
    });
    const taskCompletion = task.addResource('completion');
    taskCompletion.addMethod('PUT', apiLambdaIntegration, {
      authorizationType: apiGW.AuthorizationType.IAM
    });
  }
}
