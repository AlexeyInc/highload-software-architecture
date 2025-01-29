## Project Overview

This project demonstrates how to handle concurrent requests and database operations under stress tests. The application processes incoming requests by inserting new documents into a MongoDB collection and updating the timestamp if a document already exists.
 

### Stress tests results
1. **Concurrency 10**
```
siege -f urls.txt -c10 -t 15s
```
```
Transactions:		     9394    hits
Availability:		      100.00 %
Elapsed time:		       15.15 secs
Data transferred:	        0.55 MB
Response time:		       16.01 ms
Transaction rate:	      620.07 trans/sec
Throughput:		        0.04 MB/sec
Concurrency:		        9.93
Successful transactions:     9394
Failed transactions:	        0
Longest transaction:	      210.00 ms
Shortest transaction:	       10.00 ms
```

2. **Concurrency 25**
```
siege -f urls.txt -c25 -t 15s
```
```
Transactions:		    16384    hits
Availability:		      100.00 %
Elapsed time:		       15.62 secs
Data transferred:	        0.95 MB
Response time:		       13.91 ms
Transaction rate:	     1048.91 trans/sec
Throughput:		        0.06 MB/sec
Concurrency:		       14.59
Successful transactions:    16384
Failed transactions:	        0
Longest transaction:	     1020.00 ms
Shortest transaction:	       10.00 ms
```

3. Concurrency 50
```
siege -f urls.txt -c50 -t 15s
```
```
Transactions:		    16431    hits
Availability:		      100.00 %
Elapsed time:		       15.33 secs
Data transferred:	        0.96 MB
Response time:		       23.16 ms
Transaction rate:	     1071.82 trans/sec
Throughput:		        0.06 MB/sec
Concurrency:		       24.83
Successful transactions:    16431
Failed transactions:	        0
Longest transaction:	     6740.00 ms
Shortest transaction:	       10.00 ms
```

4. Concurrency 100
```
siege -f urls.txt -c100 -t 15s
```
```
Transactions:		    16465    hits
Availability:		      100.00 %
Elapsed time:		       15.37 secs
Data transferred:	        0.96 MB
Response time:		       32.10 ms
Transaction rate:	     1071.24 trans/sec
Throughput:		        0.06 MB/sec
Concurrency:		       34.39
Successful transactions:    16465
Failed transactions:	        0
Longest transaction:	     6740.00 ms
Shortest transaction:	       10.00 ms
```

### Summary
- **Availability**: The application exhibits excellent reliability with 100% availability across all concurrency levels.
- **Response** Time: Performance remains good at lower and moderate concurrency levels but degrades noticeably under high concurrency.
- **Throughput**: Throughput is consistent and scales with concurrency until reaching a plateau at higher levels, indicating a bottleneck in the system.
