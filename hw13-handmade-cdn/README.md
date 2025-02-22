## Project Overview

This project is a simple CDN system for image delivery, using:
- bind9 DNS Server to route clients from different regions to different load balancers.
- NGINX load balancers (`load_balancer_1` and `load_balancer_2`) to distribute requests across backend nodes.
- Go-based backend servers (`node1, node2, node3, node4`) to serve cached images.
- Go-based client applications that simulate requests from Ukraine (`client_ukraine`) and Europe (`client_europe`).
- Load testing with Siege to compare different balancing strategies.


## How to Use

1. **Start the system**

```
docker-compose up --build -d
```

2. **Requests to test DNS routing**

DNS Resolution for Ukraine Client

`docker exec -it client_ukraine nslookup cdn.local 192.168.97.5`

Expected Output:
```
Name: cdn.local
Address: 192.168.97.6
```

DNS Resolution for Europe Client

`docker exec -it client_europe nslookup cdn.local 192.168.97.5`
```
Name: cdn.local
Address: 192.168.97.7
```

3. **Image requests from different regions**

*Request from Europe Ukraine to different Load Balancers*

`docker exec -it client_europe curl http://localhost:8080/request-image`

Expected Output:
```
Request to 192.168.97.6 successful: 200 OK
```


`docker exec -it client_ukraine curl http://localhost:8080/request-image`

Expected Output:
```
Request to 192.168.97.7 successful: 200 OK
```

---

**Siege load testing**

To evaluate each load balancing strategy NGINX `upstream` configuration was modified accordingly.

| Strategy               | Transactions | Data Transferred (MB) | Throughput (MB/s) |
|------------------------|-------------|----------------------|------------------|
| Round-Robin (deafult)  | 255230      | 10.22                | 0.35             |
| Weighted (one node weight=2;)   | 252330      | 10.11                | 0.35             |
| IP Hash (ip_hash;)     | 239544      | 9.59                 | 0.33             |
| Least Connections (least_conn;)| 333118      | 13.34                | 0.45             |

**Summary:**

  - Least Connections is the most efficient strategy for high-load scenarios, ensuring optimal resource usage.
  - Round-Robin and Weighted strategies are stable and predictable but may not account for uneven traffic spikes.
  - IP Hash is less efficient due to potential imbalance but may be useful when session persistence is required.
