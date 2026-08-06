FROM dhi.io/python:3.14.6-debian13-dev@sha256:9b72c38a520f44fafa1c4a3026e9b390eb3b4967c62d38be01400ecbb0232b65
ARG TARGETARCH

ENV HELM_EXPERIMENTAL_OCI=1
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

COPY ./migrate_chart.py ./migrate_chart.sh /

RUN apt-get update && apt-get install -y --no-install-recommends curl gzip \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fsSLo /helm.tar.gz "https://get.helm.sh/helm-v3.9.1-linux-${TARGETARCH}.tar.gz" \
    && tar zxvf /helm.tar.gz \
    && mkdir -p /usr/local/bin \
    && mv "/linux-${TARGETARCH}/helm" /usr/local/bin/helm \
    && rm -rf /helm.tar.gz "/linux-${TARGETARCH}" \
    && \
    pip install --break-system-packages click==8.3.1 requests==2.32.5 pyyaml==6.0.3 && \
    chmod +x /migrate_chart.sh ./migrate_chart.py

ENTRYPOINT [ "/migrate_chart.py" ]
