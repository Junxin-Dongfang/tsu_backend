## 本机环境部署步骤

1. **创建 tsu-network 网络**
    ```bash
    docker network create tsu-network
    ```

2. **启动 docker-compose-ory**
    ```bash
    docker-compose -f docker-compose-ory.local.yml up -d
    ```

3. **启动 docker-compose-main**
    ```bash
    docker-compose -f docker-compose-main.local.yml up -d
    ```

4. **启动 docker-compose-nginx**
    ```bash
    docker-compose -f docker-compose-nginx.local.yml up -d
    ```