# Cloudflared TCP 
- This is a cloudflared-tcp cli tool to manage multiple local client-side connection.
- This avoid running access multiple TCP connection and running them in individual Terminal Tab.

## prerequisites
- For this demo, you must already setup cloudflared tunnel in your server from where you want to expose TCP connection.
- Must install `cloudflared` cli in the client system.

## Installation:
- We can install using debian file from [Github releases](https://github.com/TheSpiritMan/Cloudflared-TCP/releases).
- OR
  ```sh
  curl -sSL https://raw.githubusercontent.com/TheSpiritMan/Cloudflared-TCP/main/install.sh | bash
  ```
  OR
  ```sh
  wget -qO - https://raw.githubusercontent.com/TheSpiritMan/Cloudflared-TCP/main/install.sh | bash
  ```

- Config can be found in `/etc/cloudflared-tcp.conf`.
  
- Setup Cloudflared-tcp:
  ```sh
  cloudflared-tcp setup
  ```

- Check Status:
  ```sh
  cloudflared-tcp status
  ```

- List Expose TCP port:
  ```sh
  cloudflared-tcp list
  ```

- For Help Menu:
  ```sh
  cloudflared-tcp help
  ```

## Deploy Postgres Deployment
- For this setup, I am deploying Postgres Container in K8s cluster. It can be anything of our choice.
- Postgres Manifest file can be found [here](./Manifest-Files/postgres.yaml):
- Apply Postgres:
    ```sh
    kubectl apply -f Manifest-Files/postgres.yaml
    ```
- Verify:
  ```sh
  kubectl get po,svc
  ```
  Output:
  ```sh
  NAME                           READY   STATUS    RESTARTS   AGE
  pod/postgres-db986cfcd-4j5sv   1/1     Running   0          81s

  NAME                 TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
  service/kubernetes   ClusterIP   10.43.0.1     <none>        443/TCP    38h
  service/postgres     ClusterIP   10.43.96.80   <none>        5432/TCP   81s
  ```

- Install Postgresql Client:
  ```sh
  sudo apt update
  sudo apt install -y postgresql-client
  ```

- Test connection using ClusterIP: `10.43.96.80`:
  ```sh
  PGPASSWORD=password psql -U user -d db -h 10.43.96.80
  ```

- On success, we will ve inside postgres shell.

## Exposing Postgres DB Connection using Cloudflared Tunnel.
- We will need to add below similar hostname inside cloudlare configMap with our database connection details:
  ```sh
  - hostname: "postgres-db.example.com"
    service: tcp://postgres.default.svc.cluster.local:5432
    originRequest:
      noTLSVerify: false
  ```

- Here, we need to add DNS record `postgres-db.example.com` to `cloudflared-tunnel`.

- Cloudflare does NOT directly expose raw TCP publicly so we must connect through the Cloudflare Tunnel client (cloudflared) locally.
- Install `cloudflared` cli into system. More details can be found in this [link](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/downloads/):
  ```sh
  # Add cloudflare gpg key
  sudo mkdir -p --mode=0755 /usr/share/keyrings
  curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null

  # Add this repo to your apt repositories
  # Stable
  echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared jammy main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
  # Nightly
  echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://next.pkg.cloudflare.com/cloudflared jammy main' | sudo tee /etc/apt/sources.list.d/cloudflared.list

  # install cloudflared
  sudo apt-get update && sudo apt-get install -y cloudflared
  ```

- Exposing Postgres Service to local network, here `port` can be any from our choice:
  ```sh
  cloudflared access tcp --hostname postgres-db.example.com --url localhost:15432
  ```
> Note: We have to keep open this tab for all the connection.

- Connecting to Postgres Service:
  ```sh
  PGPASSWORD=password psql -U user -d db -h localhost -p 15432
  ```

- If we intend to expose more TCP connection then we will have to follow same process:
  - First add hostname in `cloudflared` configmap. 
  - Modify DNS Record in Cloudflare (if needed).
  - Expose TCP locally using `cloudflared` cli.
  - Connect to TCP port using application.

- Repeatedly this same command in client side is painful. So have create a cli too capable of during that.
