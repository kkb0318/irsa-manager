#!/bin/bash


TARGET_BUCKET_PREFIX=s3-echoer
AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-ap-northeast-1}
JOB_TEMPLATE=job-arm.yaml.template

# deploy s3 echoer job into k8s cluster
timestamp=$(date +%s)
TARGET_BUCKET=$ROLE_NAME-$timestamp

aws s3api create-bucket \
          --bucket $TARGET_BUCKET_PREFIX \
          --create-bucket-configuration LocationConstraint=$AWS_DEFAULT_REGION \
          --region $AWS_DEFAULT_REGION

sed -e "s/TARGET_BUCKET/${TARGET_BUCKET_PREFIX}/g" ${JOB_TEMPLATE} > s3-echoer-job.yaml

kubectl create -f s3-echoer-job.yaml

echo "The S3 bucket is $TARGET_BUCKET_PREFIX"

