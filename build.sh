#!/usr/bin/env bash

# build server
cd server
docker build --no-cache -t dichlorodiphen/server .
docker push dichlorodiphen/server
cd ..

# build client
cd client
docker build --no-cache -t dichlorodiphen/client .
docker push dichlorodiphen/client
cd ..
