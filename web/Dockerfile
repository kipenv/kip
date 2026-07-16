# Static web (landing + decrypt page) served by nginx.
#
# PUBLIC_API_URL is baked in at build time (Astro). Leave it empty when the
# API is served on the same domain under /api (the recommended setup — see
# deploy/docker-compose.prod.yml); set it only if the API lives on another
# origin.
FROM node:22-alpine AS build
ARG PUBLIC_API_URL=""
ENV PUBLIC_API_URL=${PUBLIC_API_URL}
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
