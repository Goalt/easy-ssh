FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

RUN rm -f /etc/apt/sources.list.d/yarn.list \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends gh \
	&& rm -rf /var/lib/apt/lists/*