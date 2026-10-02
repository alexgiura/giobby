FROM nginx:1.27-alpine

# The official image renders /etc/nginx/templates/*.template into conf.d at start,
# substituting only defined environment variables ($uri, $host… stay as they are).
# BACKEND_HOST is the backend's service name: "backend" locally, a giobby-specific
# name on Coolify's shared network (see docker-compose.prod.yml).
ENV BACKEND_HOST=backend
COPY default.conf.template /etc/nginx/templates/default.conf.template
COPY index.html /usr/share/nginx/html/index.html

EXPOSE 80
