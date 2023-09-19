FROM node:alpine as build

WORKDIR /usr/src/client

COPY package.json .
RUN npm install

COPY . .
RUN npm run build

FROM nginx

COPY --from=build /usr/src/client/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
