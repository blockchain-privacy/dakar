
# build stage
FROM node:lts-alpine as build-stage
WORKDIR /app
COPY package*.json yarn.lock /app/
RUN yarn install
COPY . /app
RUN yarn build


# production stage
FROM nginx:stable-alpine as production-stage
WORKDIR /app
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build-stage /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

