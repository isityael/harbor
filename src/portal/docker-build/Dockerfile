FROM dhi.io/node:26.5.0-debian13-dev@sha256:728bf5fbed66fdcbe3474b5fb343a1480ee093dfd4da6dd5afaf5337ac66150c AS builder

WORKDIR /build_dir

RUN npm install --global --force npm@12.0.1

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

FROM dhi.io/nginx:1.31.3-debian13-dev@sha256:c0f34463d05737646ba965de939e228e6af7a90b4d84ceba3d5634e70740df45

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
