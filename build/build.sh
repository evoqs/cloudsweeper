#!/bin/bash

while getopts ":v:" opt; do
  case ${opt} in
    v )
      IMAGE_VERSION=$OPTARG
      ;;
    \? )
      echo "Usage: $0 -v <image_version>"
      exit 1
      ;;
    : )
      echo "Invalid option: $OPTARG requires an argument" 1>&2
      exit 1
      ;;
  esac
done
shift $((OPTIND -1))

if [ -z "$IMAGE_VERSION" ]; then
    echo "Error: Image version (-v) argument is required."
    echo "Usage: $0 -v <image_version>"
    exit 1
fi

CURRENT_DIR=$(pwd)

if [ -d "ui/cloudsweeper-ui" ]; then
    echo "Performing git pull for cloudsweeper-ui"
    cd ui/cloudsweeper-ui || exit
    git pull
    cd "$CURRENT_DIR" || exit
else
    echo "Cloning the cloudsweeper-ui repository"
    git clone -b dev git@bitbucket.org:cloudsweeper/cloudsweeper-ui.git ui/cloudsweeper-ui || exit
fi

# Remove the directory if it already exists
if [ -d "backend/cloudsweep" ]; then
    echo "Performing git pull for cloudsweeper"
    cd backend/cloudsweep || exit
    git pull
    cd "$CURRENT_DIR" || exit
    #rm -rf cloudsweep
else
   echo "Cloning the cloudsweeper repository"
   git clone -b poc git@bitbucket.org:cloudsweeper/cloudsweep.git backend/cloudsweep || exit
fi

cd backend/cloudsweep/go || exit
go build cloudsweep.go

cd "$CURRENT_DIR" || exit
# Build the Docker image
docker build -t "867226238913.dkr.ecr.us-east-1.amazonaws.com/cs-ui:$IMAGE_VERSION" ./ui || exit
docker build -t "867226238913.dkr.ecr.us-east-1.amazonaws.com/cs:$IMAGE_VERSION" ./backend || exit

#ECR_PASSWORD=$(aws ecr get-login-password --region us-east-1)
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 867226238913.dkr.ecr.us-east-1.amazonaws.com
docker push 867226238913.dkr.ecr.us-east-1.amazonaws.com/cs:$IMAGE_VERSION
docker push 867226238913.dkr.ecr.us-east-1.amazonaws.com/cs-ui:$IMAGE_VERSION

#docker compose -f ./cs_compose.yaml up -d
