FROM node:22-alpine AS build
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build -- --configuration production

FROM nginx:1.27-alpine
COPY --from=build /app/dist/wealthos-web/browser /usr/share/nginx/html
EXPOSE 80

