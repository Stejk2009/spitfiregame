FROM mcr.microsoft.com/vscode/devcontainers/go:0-1.17

# Install any additional dependencies here
RUN apt-get update && apt-get install -y curl gnupg && \
    curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor -o /usr/share/keyrings/yarnkey.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/yarnkey.gpg] https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list && \
    apt-key del 23E7166788B63E1E && \
    apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog
