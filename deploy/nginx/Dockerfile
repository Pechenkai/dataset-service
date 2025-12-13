FROM nginx:1.25-alpine

COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY static /usr/share/nginx/html/static
COPY internal/delivery/web/static /usr/share/nginx/html/static
COPY readme.md /usr/share/nginx/html/readme.md
RUN mkdir -p /var/cache/nginx/ppo

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1/status || exit 1
