FROM dhi.io/node:26.11.1-debian13-dev@sha256:45c70bf430695765dc3227298e548c6c393e3a53b92aaf1a7f8389c22902bce0 AS builder

WORKDIR /build_dir

RUN npm install --global --force npm@12.2.0

COPY src/portal /build_dir
COPY api/v2.0/swagger.yaml /build_dir/swagger.yaml


RUN apt-get update \
    && apt-get install -y --no-install-recommends python3-yaml \
    && rm -rf /var/lib/apt/lists/*
RUN npm ci
RUN npm run postinstall
RUN npm run generate-build-timestamp
RUN node --max_old_space_size=2048 'node_modules/@angular/cli/bin/ng' build --configuration production
RUN python3 -c 'import sys, yaml, json; y=yaml.safe_load(sys.stdin.read()); print(json.dumps(y))' < swagger.yaml > dist/swagger.json
RUN cd app-swagger-ui && npm ci
RUN cd app-swagger-ui && npm run build

FROM dhi.io/nginx:1.31.6-debian13-dev@sha256:8f8a690387d75205568103d234f2b6a348f6ee8183708de0e501cc3f8a5d5df0

USER 0
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*


COPY --from=builder /build_dir/dist /usr/share/nginx/html
COPY --from=builder /build_dir/app-swagger-ui/dist /usr/share/nginx/html
COPY src/portal/docker-build/nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080
VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]
