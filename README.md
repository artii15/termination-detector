# Termination detector
Service responsible for detecting termination of processes consisting of multiple tasks 
which potentially can be executed in parallel. The service provides a RESTful API which
allows registering new tasks, indicating tasks completion results and checking process 
termination status.

The service comes with SDK which simplifies interaction with the API.
There is also an AWS CDK deployment code allowing simple service deployment in AWS.

## Building and deploying
Simply type `make`. The command will build service binary and package it into zip file.
In order to deploy the service, run `npx cdk deploy`.

