# Load Balancer with GeoIP2 using NGINX 

## Project Overview

This project sets up a load balancer using NGINX with GeoIP2 to route traffic based on the user's location. The setup includes:

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

"here screeenshot example will be"

4. Use Touch VPN Chrome Extension to simulate requests from different regions (DE, GB, etc.).

## Application in Action

After setting up ngrok, you can use the generated external URL to test routing behavior. 
Requests will be forwarded to the appropriate servers based on location. The `X-Debug-Country-Code` header can be checked to verify which country was detected.

1. Request from 

