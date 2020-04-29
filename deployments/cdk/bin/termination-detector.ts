#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { TerminationDetectorStack } from '../lib/termination-detector-stack';

const app = new cdk.App();
new TerminationDetectorStack(app, 'TerminationDetectorStack');
