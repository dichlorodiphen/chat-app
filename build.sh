#!/usr/bin/env bash

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
