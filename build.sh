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

# wait for database to deploy
echo 'waiting for database to deploy'
sleep 20s
echo 'database deployed, now streaming backend logs'

# stream logs from backend
pods=$(minikube kubectl -- get pods)
pattern='(backend[0-9a-z-]+)'
if [[ $pods =~ $pattern ]]; then
  minikube kubectl -- logs -f ${BASH_REMATCH[0]}
else
  echo "unexpected error"
fi

