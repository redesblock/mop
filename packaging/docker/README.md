# Docker compose

The docker-compose provides an app container for MOP itself.
To prepare your machine to run docker compose execute
```
mkdir -p mop && cd mop
wget -q https://raw.githubusercontent.com/redesblock/mop/master/packaging/docker/docker-compose.yml
wget -q https://raw.githubusercontent.com/redesblock/mop/master/packaging/docker/env -O .env
```
Set all configuration variables inside `.env`

If you want to run node in full mode, set `MOP_FULL_NODE=true`

MOP requires an BNB Smart Chain endpoint to function. Obtain a free Infura account and set:
- `MOP_BSC_SWAP_ENDPOINT=wss://bnb.infura.io/ws/v3/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

Set mop password by either setting `MOP_PASSWORD` or `MOP_PASSWORD_FILE`

If you want to use password file set it to
- `MOP_PASSWORD_FILE=/password`

Mount password file local file system by adding
```
- ./password:/password
```
to mop volumes inside `docker-compose.yml`

Start it with
```
docker-compose up -d
```

From logs find URL line with `on bnb smart chain you can get both bnb and mop from` and prefund your node
```
docker-compose logs -f mop-1
```

Update services with
```
docker-compose pull && docker-compose up -d
```

## Running multiple Mop nodes
It is easy to run multiple mop nodes with docker compose by adding more services to `docker-compose.yaml`
To do so, open `docker-compose.yaml`, copy lines 3-58 and past this after line 58.
In the copied lines, replace all occurences of `mop-1` with `mop-2` and adjust the `API_ADDR` and `P2P_ADDR` and `DEBUG_API_ADDR` to respectively `1933`, `1934` and `127.0.0.1:1935`
Lastly, add your newly configured services under `volumes` (last lines), such that it looks like:
```yaml
volumes:
  mop-1:
  mop-2:
```

If you want to create more than two nodes, simply repeat the process above, ensuring that you keep unique name for your mop and clef services and update the ports
