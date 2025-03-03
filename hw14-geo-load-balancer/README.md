# Load Balancer with GeoIP2 using NGINX 

## Project Overview

This project sets up a load balancer using NGINX with GeoIP2 to route traffic based on the user's location. 

The setup includes:
- A load balancer (NGINX) with GeoIP2 to determine the request origin.

- Four main servers:

    - One server for requests from Germany (DE).

    - Two servers handling requests from Great Britain (GB) in a round-robin fashion.

    - One server for all other locations (worldwide traffic).

    - A backup server that receives traffic in case of failures.

GeoIP2 database (`GeoLite2-Country.mmdb`) for country-based traffic routing.

Health checks are performed every 5 seconds, and failed servers route traffic to the backup server.


## How to Use

1. Start the services using Docker Compose: `docker-compose up -d`

2. Check that server is responding 
```
curl -I localhost:80 
                              
HTTP/1.1 200 OK
```

3. Expose the local application to the external world using `ngrok`. 
Run command: `ngrok http http://localhost:80`

This should create a new session with an external URL.

![Screenshot 2025-03-03 at 14 33 15](https://github.com/user-attachments/assets/a59e5c9b-ea01-4c0e-8e8f-a041f334af41)


4. Use Touch VPN Chrome Extension to simulate requests from different regions (DE, GB, etc.).

## Application in Action

After setting up ngrok, you can use the generated external URL to test routing behavior. 
Requests will be forwarded to the appropriate servers based on location. The `X-Debug-Country-Code` header can be checked to verify which country was detected.

Scenario below validates the load balancing mechanism across regional servers and ensures failover functionality when primary servers are unavailable.

1. Request from Germany (DE)

<img width="920" alt="Screenshot 2025-03-03 at 14 08 36" src="https://github.com/user-attachments/assets/ef0e71c5-404e-4b3b-8e85-caf1067fcc15" />

**Server 1** (`de_server`) is triggered and logs the request.

![Screenshot 2025-03-03 at 14 08 43](https://github.com/user-attachments/assets/29125314-8a42-4866-97d2-ff53e3763b54)

2.  Multiple Requests from the United Kingdom (GB)

<img width="890" alt="Screenshot 2025-03-03 at 14 14 24" src="https://github.com/user-attachments/assets/fb16e484-4056-4dd2-84e7-aef1f3cb1f97" />

**Server 2 & Server 3** (`gb_server_1` & `gb_server_1`) handle the traffic and log the requests.

![Screenshot 2025-03-03 at 14 17 13](https://github.com/user-attachments/assets/84d0f781-bb21-4e5b-9927-abcb53843a73)

3. Request from the United States (US)

<img width="927" alt="Screenshot 2025-03-03 at 14 17 24" src="https://github.com/user-attachments/assets/584405ad-144f-4bd1-800d-35825255b7be" />


**Server 4**  (`world_server`)  is triggered and logs the request.

![Screenshot 2025-03-03 at 14 01 01](https://github.com/user-attachments/assets/96870474-6b27-4f3a-a4d8-52a4ad553b32)

4. Testing Failover to backup_server
   
    - All four servers (`de_server, gb_server_1, gb_server_2, and world_server`) are shut down.
    - A request is made from any region.
    - The request initially fails within the `fail_timeout` window (~5s).
    - The `backup_server` takes over and logs the request.

![Screenshot 2025-03-03 at 15 11 41](https://github.com/user-attachments/assets/db2e5f99-c54f-48f4-927b-db090590f60b)
