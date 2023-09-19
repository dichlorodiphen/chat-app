#!/usr/bin/env bash

# delete resources
minikube kubectl -- delete -Rf deployment/

# build server
cd server
docker build -t dichlorodiphen/server .
docker push dichlorodiphen/server
cd ..

# build client
cd client
docker build -t dichlorodiphen/client .
docker push dichlorodiphen/client
cd ..

# redeploy
minikube kubectl -- apply -Rf deployment/
minikube kubectl -- get pods

